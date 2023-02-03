// Package reggen generates text based on regex definitions
package reggen

import (
	"math"
	"math/rand"
	"regexp/syntax"
)

const runeRangeEnd = 0x10ffff
const printableChars = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ!\"#$%&'()*+,-./:;<=>?@[\\]^_`{|}~ \t\n\r"

var printableCharsNoNL = printableChars[:len(printableChars)-2]

type state struct {
	limit int
}

func generate(s *state, re *syntax.Regexp, rand *rand.Rand) string {
	op := re.Op
	switch op {
	case syntax.OpNoMatch:
	case syntax.OpEmptyMatch:
		return ""
	case syntax.OpLiteral:
		res := ""
		for _, r := range re.Rune {
			res += string(r)
		}
		return res
	case syntax.OpCharClass:
		// number of possible chars
		sum := 0
		for i := 0; i < len(re.Rune); i += 2 {
			sum += int(re.Rune[i+1]-re.Rune[i]) + 1
			if re.Rune[i+1] == runeRangeEnd {
				sum = -1
				break
			}
		}
		// pick random char in range (inverse match group)
		if sum == -1 {
			possibleChars := []uint8{}
			for j := 0; j < len(printableChars); j++ {
				c := printableChars[j]
				//fmt.Printf("Char %c %d\n", c, c)
				// Check c in range
				for i := 0; i < len(re.Rune); i += 2 {
					if rune(c) >= re.Rune[i] && rune(c) <= re.Rune[i+1] {
						possibleChars = append(possibleChars, c)
						break
					}
				}
			}
			//fmt.Println("Possible chars: ", possibleChars)
			if len(possibleChars) > 0 {
				c := possibleChars[rand.Intn(len(possibleChars))]
				return string([]byte{c})
			}
		}
		r := rand.Intn(sum)
		var ru rune
		sum = 0
		for i := 0; i < len(re.Rune); i += 2 {
			gap := int(re.Rune[i+1]-re.Rune[i]) + 1
			if sum+gap > r {
				ru = re.Rune[i] + rune(r-sum)
				break
			}
			sum += gap
		}
		return string(ru)
	case syntax.OpAnyCharNotNL, syntax.OpAnyChar:
		chars := printableChars
		if op == syntax.OpAnyCharNotNL {
			chars = printableCharsNoNL
		}
		c := chars[rand.Intn(len(chars))]
		return string([]byte{c})
	case syntax.OpBeginLine:
	case syntax.OpEndLine:
	case syntax.OpBeginText:
	case syntax.OpEndText:
	case syntax.OpWordBoundary:
	case syntax.OpNoWordBoundary:
	case syntax.OpCapture:
		return generate(s, re.Sub0[0], rand)
	case syntax.OpStar:
		// Repeat zero or more times
		res := ""
		count := rand.Intn(s.limit + 1)
		for i := 0; i < count; i++ {
			for _, r := range re.Sub {
				res += generate(s, r, rand)
			}
		}
		return res
	case syntax.OpPlus:
		// Repeat one or more times
		res := ""
		count := rand.Intn(s.limit) + 1
		for i := 0; i < count; i++ {
			for _, r := range re.Sub {
				res += generate(s, r, rand)
			}
		}
		return res
	case syntax.OpQuest:
		// Zero or one instances
		res := ""
		count := rand.Intn(2)

		for i := 0; i < count; i++ {
			for _, r := range re.Sub {
				res += generate(s, r, rand)
			}
		}
		return res
	case syntax.OpRepeat:
		// Repeat one or more times
		res := ""
		count := 0
		re.Max = int(math.Min(float64(re.Max), float64(s.limit)))
		if re.Max > re.Min {
			count = rand.Intn(re.Max - re.Min + 1)
		}

		for i := 0; i < re.Min || i < (re.Min+count); i++ {
			for _, r := range re.Sub {
				res += generate(s, r, rand)
			}
		}
		return res
	case syntax.OpConcat:
		// Concatenate sub-regexes
		res := ""
		for _, r := range re.Sub {
			res += generate(s, r, rand)
		}
		return res
	case syntax.OpAlternate:
		i := rand.Intn(len(re.Sub))
		return generate(s, re.Sub[i], rand)
	default:

	}
	return ""
}

// limit is the maximum number of times star, range or plus should repeat
// i.e. [0-9]+ will generate at most 10 characters if this is set to 10
func Generate(re *syntax.Regexp, limit int, rand *rand.Rand) string {
	return generate(&state{limit: limit}, re, rand)
}

func GenerateFromString(re string, limit int, rand *rand.Rand) (string, error) {
	regex, err := syntax.Parse(re, syntax.Perl)
	if err != nil {
		return "", err
	}
	return Generate(regex, limit, rand), nil
}
