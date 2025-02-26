package encode

import "github.com/codeharik/secretary/utils"

var SEC32KeyMap = [][]byte{
	/*00*/ {'~'},
	/*01*/ {'a', 'A'},
	/*02*/ {'p', 'P', 'b', 'B', '+'},
	/*03*/ {'c', 'C', 's', 'S', ':'},
	/*04*/ {'d', 'D', '/'},
	/*05*/ {'e', 'E', '='},
	/*06*/ {'i', 'I'},
	/*07*/ {'f', 'F', '"', '\''},
	/*08*/ {'g', 'G', 'j', 'J', 'x', 'X', 'z', 'Z'},
	/*09*/ {'h', 'H', '%'},
	/*10*/ {'o', 'O'},
	/*11*/ {'k', 'K', 'q', 'Q', '?', '!', '\\', '|'},
	/*12*/ {'l', 'L', '&', '@'},
	/*13*/ {'m', 'M', '-'},
	/*14*/ {'n', 'N', '*'},
	/*15*/ {'t', 'T'},
	/*16*/ {'r', 'R', '^'},
	/*17*/ {'u', 'U', 'v', 'V', '#'},
	/*18*/ {'w', 'W', 'y', 'Y', '$'},
	/*19*/ {'0'},
	/*20*/ {'1'},
	/*21*/ {'2'},
	/*22*/ {'3'},
	/*23*/ {'4'},
	/*24*/ {'5'},
	/*25*/ {'6'},
	/*26*/ {'7'},
	/*27*/ {'8'},
	/*28*/ {'9'},
	/*29*/ {'(', '[', '<', '{'},
	/*30*/ {')', ']', '>', '}'},
	/*31*/ {'_', '.', ',', ';', ' ', '\n'},
}

const SEC32 = `-apcdeifghoklmntruw0123456789QR_`

var (
	ASCII32Index = [256]byte{}
	SEC32Index   = [256]byte{}
)

func init() {
	for i := range ASCII32Index {
		ASCII32Index[i] = 0
		SEC32Index[i] = 0
	}
	for index, chars := range SEC32KeyMap {
		SEC32Index[SEC32[index]] = byte(index)
		for _, c := range chars {
			ASCII32Index[c] = byte(index)
		}
	}
}

func AsciiToSec32(str string) string {
	return string(utils.Map(
		[]byte(str),
		func(c byte) byte {
			return SEC32[ASCII32Index[c]]
		}))
}

func Ascii32ToIndex(str string) []byte {
	return utils.Map(
		[]byte(str),
		func(c byte) byte {
			return ASCII32Index[c]
		})
}

func IndexToAscii32(arr []byte) string {
	return string(utils.Map(
		arr,
		func(i byte) byte {
			return SEC32KeyMap[i][0]
		}))
}

func Sec32ToAscii(str string) string {
	return string(utils.Map(
		[]byte(str),
		func(c byte) byte {
			return SEC32KeyMap[SEC32Index[c]][0]
		}))
}

func Sec32ToIndex(str string) []byte {
	return utils.Map(
		[]byte(str),
		func(c byte) byte {
			return SEC32Index[c]
		})
}

func IndexToSec32(arr []byte) string {
	return string(utils.Map(
		arr,
		func(i byte) byte {
			return SEC32[i]
		}))
}
