package libs

import (
	"github.com/fatih/color"
)

// Banner print ascii banner
func Banner() string {
	version := color.HiWhiteString(VERSION)
	author := color.MagentaString(AUTHOR)
	b := color.GreenString(``)

	b += "\n" + color.HiGreenString(``)
	b += "\n" + color.GreenString(` ██████╗  ██████╗ ███████╗ ██████╗██╗  ██╗███╗   ██╗ `)
	b += "\n" + color.GreenString(`██╔════╝ ██╔═══██╗██╔════╝██╔════╝██║  ██║████╗  ██║ `)
	b += "\n" + color.GreenString(`██║  ███╗██║   ██║███████╗██║     ███████║██╔██╗ ██║'                              `)
	b += "\n" + color.GreenString(`██║   ██║██║   ██║╚════██║██║     ╚════██║██║╚██╗██║`)
	b += "\n" + color.GreenString(`╚██████╔╝╚██████╔╝███████║╚██████╗     ██║██║ ╚████║`)
	b += "\n" + color.GreenString(` ╚═════╝  ╚═════╝ ╚══════╝ ╚═════╝     ╚═╝╚═╝  ╚═══╝`)
	b += "\n" + color.GreenString(``)
	b += "\n" + color.CyanString(`         		 🚀 goSc4n %v`, version) + color.CyanString(` by %v 🚀`, author)
	b += "\n\n" + color.HiWhiteString(`               The Web Application Security Scanner`)
	b += "\n\n" + color.HiGreenString(`                                     ¯\_(ツ)_/¯`) + "\n\n"
	color.Unset()
	return b
}
