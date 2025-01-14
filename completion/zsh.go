package completion

import (
	"fmt"
	"io"
	"strings"

	"github.com/alecthomas/kong"
)

// Zsh is zsh completion generator.
type Zsh struct{}

// Run generates zsh completion script.
func (z Zsh) Run(ctx *kong.Context) error {
	var out strings.Builder
	format := `#compdef %[1]s
# zsh completion for %[1]s
# generated by gum completion

`
	fmt.Fprintf(&out, format, ctx.Model.Name)
	z.gen(&out, ctx.Model.Node)
	_, err := fmt.Fprint(ctx.Stdout, out.String())
	return err
}

func (z Zsh) writeFlag(buf io.StringWriter, f *kong.Flag) {
	var str strings.Builder
	str.WriteString("        ")
	if f.Short != 0 {
		str.WriteString("'(")
		str.WriteString(fmt.Sprintf("-%c --%s", f.Short, f.Name))
		if !f.IsBool() {
			str.WriteString("=")
		}
		str.WriteString(")'")
		str.WriteString("{")
		str.WriteString(fmt.Sprintf("-%c,--%s", f.Short, f.Name))
		if !f.IsBool() {
			str.WriteString("=")
		}
		str.WriteString("}")
		str.WriteString("\"")
	} else {
		str.WriteString("\"")
		str.WriteString(fmt.Sprintf("--%s", f.Name))
		if !f.IsBool() {
			str.WriteString("=")
		}
	}
	str.WriteString(fmt.Sprintf("[%s]", f.Help))
	if !f.IsBool() {
		str.WriteString(":")
		str.WriteString(strings.ToLower(f.Help))
		str.WriteString(":")
	}
	values := flagPossibleValues(f)
	if len(values) > 0 {
		str.WriteString("(")
		for i, v := range f.EnumSlice() {
			str.WriteString(v)
			if i < len(values)-1 {
				str.WriteString(" ")
			}
		}
		str.WriteString(")")
	}
	str.WriteString("\"")
	writeString(buf, str.String())
}

func (z Zsh) writeFlags(buf io.StringWriter, cmd *kong.Node) {
	for i, f := range cmd.Flags {
		if f.Hidden {
			continue
		}
		z.writeFlag(buf, f)
		if i < len(cmd.Flags)-1 {
			writeString(buf, " \\\n")
		}
	}
}

func (z Zsh) writeCommand(buf io.StringWriter, c *kong.Node) {
	writeString(buf, fmt.Sprintf("                \"%s[%s]\"", c.Name, c.Help))
}

func (z Zsh) writeCommands(buf io.StringWriter, cmd *kong.Node) {
	for i, c := range cmd.Children {
		if c == nil || c.Hidden {
			continue
		}
		z.writeCommand(buf, c)
		if i < len(cmd.Children)-1 {
			_, _ = buf.WriteString(" \\")
		}
		writeString(buf, "\n")
	}
}

func (z Zsh) gen(buf io.StringWriter, cmd *kong.Node) {
	for _, c := range cmd.Children {
		if c == nil || c.Hidden {
			continue
		}
		z.gen(buf, c)
	}
	cmdName := commandName(cmd)

	writeString(buf, fmt.Sprintf("_%s() {\n", cmdName))
	if hasCommands(cmd) {
		writeString(buf, "    local line state\n")
	}
	writeString(buf, "    _arguments -C \\\n")
	z.writeFlags(buf, cmd)
	if hasCommands(cmd) {
		writeString(buf, " \\\n")
		writeString(buf, "        \"1: :->cmds\" \\\n")
		writeString(buf, "        \"*::arg:->args\"\n")
		writeString(buf, "    case \"$state\" in\n")
		writeString(buf, "        cmds)\n")
		writeString(buf, fmt.Sprintf("            _values \"%s command\" \\\n", cmdName))
		z.writeCommands(buf, cmd)
		writeString(buf, "            ;;\n")
		writeString(buf, "        args)\n")
		writeString(buf, "            case \"$line[1]\" in\n")
		for _, c := range cmd.Children {
			if c == nil || c.Hidden {
				continue
			}
			writeString(buf, fmt.Sprintf("                %s)\n", c.Name))
			writeString(buf, fmt.Sprintf("                    _%s\n", commandName(c)))
			writeString(buf, "                    ;;\n")
		}
		writeString(buf, "            esac\n")
		writeString(buf, "            ;;\n")
		writeString(buf, "    esac\n")
	}
	// writeArgAliases(buf, cmd)
	writeString(buf, "\n")
	writeString(buf, "}\n\n")
}
