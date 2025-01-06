package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"strings"

	"github.com/mohprilaksono/c.go/app"
)

func main() {
	// platforms := []string{"python3", "fasm-x86_64-linux"}

	input := os.Args[1]
	filePath := input
	source, _ := os.ReadFile(filepath.Join("f", filePath))
	lexer := app.Lexer{
		FilePath: filePath,
		Source: string(source),
	}
	fn, found := parseFunc(lexer)
	if !found {
		os.Exit(1)
	}

	generatePython3(fn)

	//     $file_path = $input;
//     $source = file_get_contents($file_path);
//     if (!$source) exit(69);
//     $lexer = new Lexer($file_path, $source);
//     $func = parse_function($lexer);
//     if (!$func) exit(69);

//     switch ($platform) {
//     case "python3": generate_python3($func); break;
//     case "fasm-x86_64-linux": generate_fasm_x86_64_linux($func); break;
//     default: todo("unreachable"); break;
//     }
}

// function main($argv) {
//     $platforms = array(
//         "python3",
//         "fasm-x86_64-linux"
//     );

//     $program = array_shift($argv);
//     $input = null;
//     $platform = $platforms[0];

//     while (sizeof($argv) > 0) {
//         $flag = array_shift($argv);
//         switch ($flag) {
//         case "-help": {
//             usage($program);
//             exit(0);
//         } break;

//         case "-target": {
//             if (sizeof($argv) === 0) {
//                 usage($program);
//                 print("ERROR: no value was provided for flag $flag\n");
//                 exit(69);
//             }

//             $arg = array_shift($argv);

//             if ($arg === "list") {
//                 print("Available targets:\n");
//                 foreach ($platforms as $p) {
//                     print("    $p\n");
//                 }
//                 exit(69);
//             }

//             if (in_array($arg, $platforms)) {
//                 $platform = $arg;
//             } else {
//                 usage($program);
//                 print("ERROR: unknown target $arg\n");
//                 exit(69);
//             }
//         } break;

//         default:
//             $input = $flag;
//         }
//     }

//     if ($input === null) {
//         usage($program);
//         print("ERROR: no input is provided\n");
//         exit(69);
//     }

//     $file_path = $input;
//     $source = file_get_contents($file_path);
//     if (!$source) exit(69);
//     $lexer = new Lexer($file_path, $source);
//     $func = parse_function($lexer);
//     if (!$func) exit(69);

//     switch ($platform) {
//     case "python3": generate_python3($func); break;
//     case "fasm-x86_64-linux": generate_fasm_x86_64_linux($func); break;
//     default: todo("unreachable"); break;
//     }
// }

func expectToken(lexer app.Lexer, types ...string) (app.Token, bool) {
	token, success := lexer.NextToken()
	if success {
		fmt.Printf("%s: ERROR: expected %s but got end of file\n", 	lexer.Loc().Display(), strings.Join(types, " or "))
		return app.Token{}, false
	}

	for _, t := range types {
		if token.Type == t {
			return token, true
		}
	}

	fmt.Printf("%s: ERROR: expected %s but got %s\n", lexer.Loc().Display(), strings.Join(types, " or "), token.Type)
	return app.Token{}, false
}

func parseType(lexer app.Lexer) string {
	token, _ := expectToken(lexer, app.TOKEN_NAME)
	if (fmt.Sprint(token.Value)) != "int" {
		fmt.Printf("%s: ERROR: unexpected type %s", token.Loc.Display(), token.Value)
		return ""
	}

	return app.TYPE_INT
}

func parseArglist(lexer app.Lexer) []string {	
	_, found := expectToken(lexer, app.TOKEN_OPAREN)
	if !found {
		return nil
	}

	var argLists []string

	// First argument (optional).
	expr, found := expectToken(lexer, app.TOKEN_STRING, app.TOKEN_NUMBER, app.TOKEN_CPAREN)
	if !found {
		return nil
	}

	if expr.Type == app.TOKEN_CPAREN {
		// Call with no arguments.
		return argLists
	}

	argLists = append(argLists, fmt.Sprint(expr.Value))

	// Second, third, etc. arguments (optional).
	for {
		expr, found := expectToken(lexer, app.TOKEN_CPAREN, app.TOKEN_COMMA)
		if !found {
			return nil
		}
		
		if expr.Type == app.TOKEN_CPAREN {
			break
		}

		expr, found = expectToken(lexer, app.TOKEN_STRING, app.TOKEN_NUMBER)
		if !found {
			return nil
		}

		argLists = append(argLists, fmt.Sprintf("%v", expr.Value))
	}

	return argLists
}

func parseBlock(lexer app.Lexer) []any {
	if _, found := expectToken(lexer, app.TOKEN_OCURLY); !found {
		return nil
	}

	var block []any

	for {
		name, found := expectToken(lexer, app.TOKEN_NAME, app.TOKEN_CCURLY)
		if !found {
			return nil
		}

		if name.Type == app.TOKEN_CCURLY {
			break
		}

		if (fmt.Sprint(name.Value)) == "return" {
			expr, found := expectToken(lexer, app.TOKEN_NUMBER, app.TOKEN_STRING)
			if !found {
				return nil
			}

			block = append(block, app.RetStmt{Expr: expr.Value})
		} else {
			argList := parseArglist(lexer)
			if argList == nil {
				return nil
			}

			block = append(block, app.FuncallStmt{Name: name, Args: argList})
		}

		if _, found := expectToken(lexer, app.TOKEN_SEMICOLON); !found {
			return nil
		}
	}

	return block
}

func parseFunc(lexer app.Lexer) (app.Func, bool) {
	t := parseType(lexer)
	if t == "" {
		return app.Func{}, false
	}

	name, found := expectToken(lexer, app.TOKEN_NAME)
	if !found {
		return app.Func{}, false
	}

	if _, found := expectToken(lexer, app.TOKEN_OPAREN); !found {
		return app.Func{}, false
	}

	if _, found := expectToken(lexer, app.TOKEN_CPAREN); !found {
		return app.Func{}, false
	}

	body := parseBlock(lexer)

	return app.Func{
		Name: name,
		Body: body,
	}, true
}

func literalToPy(value any) string {
	if reflect.TypeOf(value).Name() == "string" {
		return fmt.Sprintf("\"%s\"", strings.ReplaceAll(value.(string), "\n", "\\n"))
	}
	
	return fmt.Sprint(value)
}

func generatePython3(fn app.Func) {
	for _, stmt := range fn.Body {
		if reflect.TypeOf(stmt).Name() == "FuncallStmt" {
			statement := stmt.(*app.FuncallStmt)
			stmtName := statement.Name.(*app.Token)
			if (fmt.Sprint(stmtName.Value)) == "printf" {
				format := statement.Args[0]
				if len(statement.Args) <= 1 {
					if strings.HasSuffix(format, "\\n") {
						formatWithoutNewline := format[:len(format) - 2]
						fmt.Printf("print(%s)\n", literalToPy(formatWithoutNewline)) 
					} else {
						fmt.Printf("print(%s, end=\"\")\n", literalToPy(format))
					}
				} else {
					substitutions := new(strings.Builder)

					substitutions.WriteString(" % (")
					for i, arg := range statement.Args {
						if i == 0 {
							continue
						}

						substitutions.WriteString(literalToPy(arg))
						substitutions.WriteRune(',')
					}

					substitutions.WriteByte(')')
					fmt.Printf("print(%s%s, end=\"\")\n", literalToPy(format), substitutions.String())
				}
			} else {
				fmt.Printf("%s: ERROR: unknown function %s\n", stmtName.Loc.Display(), stmtName.Value)
				os.Exit(1)
			}
		}
	}
}

func generateFasmX8664Linux(fn app.Func) {
	out := os.Stdout
	defer out.Close()

	out.WriteString("format ELF64 executable 3\n")
	out.WriteString("segment readable executable\n")
	out.WriteString("entry start\n")
	out.WriteString("start:\n")

	var str []string

	for _, stmt := range fn.Body {
		if reflect.TypeOf(stmt).Name() == "RetStmt" {
			statement := stmt.(*app.RetStmt)

			out.WriteString("    mov rax, 60\n")
			fmt.Fprintf(out, "    mov rdi, %v\n", statement.Expr)
			out.WriteString("    syscall\n")
		} else if reflect.TypeOf(stmt).Name() == "FuncallStmt" {
			statement := stmt.(*app.FuncallStmt)
			token := statement.Name.(*app.Token)
			if (fmt.Sprint(token.Value)) == "printf" {
				arity := len(statement.Args)
				if arity != 1 {
					fmt.Fprintf(out, "%s: ERROR: expected 1 argument but, got %d\n", token.Loc.Display(), arity)
					os.Exit(1)
				}

				format := statement.Args[0]
				n := len(str)
				m := len(format)
				out.WriteString("    mov rax, 1\n")
				out.WriteString("    mov rdi, 1\n")
				fmt.Fprintf(out, "    mov rsi, str_%d\n", n)
				fmt.Fprintf(out, "    mov rdx, %d\n", m)
				io.WriteString(out, "    syscall\n")

				str = append(str, format)
			} else {
				fmt.Fprintf(out, "%s: ERROR: unknown function %s\n", token.Loc.Display(), token.Value)
				os.Exit(1)
			}
		} else {
			panic("unreachable")
		}
	}

	fmt.Fprintln(out, "segment readable writable")
	for n, s := range str {
		fmt.Fprintf(out, "str_%d db ", n)
		m := len(str)
		for i := 0; i < m; i++ {
			c := s[i]
			if i > 0 {
				out.Write([]byte(","))
			}

			fmt.Fprint(out, c)
		}

		fmt.Fprintln(out)
	}
}

func usage(program string) {
	out := os.Stdout
	defer out.Close()

	fmt.Fprintf(out, "Usage: php %s [OPTIONS] <input.c>\n", program)
	fmt.Fprintln(out, "OPTIONS:")
	fmt.Fprintln(out, "    -target <target>    Compilation target. Provide `list` to get the list of targets. (default: python3)")
	fmt.Fprintln(out, "    -help               Print this message")
}
