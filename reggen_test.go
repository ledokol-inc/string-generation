package reggen

import (
	"fmt"
	"math/rand"
	"regexp"
	"regexp/syntax"
	"testing"
	"time"
)

type testCase struct {
	regex string
}

var cases = []testCase{
	{`123[0-2]+.*\w{3}`},
	{`^\d{1,2}[/](1[0-2]|[1-9])[/]((19|20)\d{2})$`},
	{`^((25[0-5]|(2[0-4]|1{0,1}[0-9]){0,1}[0-9])\.){3}(25[0-5]|(2[0-4]|1{0,1}[0-9]){0,1}[0-9])$`},
	{`^\d+$`},
	{`\D{3}`},
	{`((123)?){3}`},
	{`(ab|bc)def`},
	{`[^abcdef]{5}`},
	{`[^1]{3,5}`},
	{`[[:upper:]]{5}`},
	{`[^0-5a-z\s]{5}`},
	{`Z{2,5}`},
	{`[a-zA-Z]{100}`},
	{`^[a-z]{5,10}@[a-z]{5,10}\.(com|net|org)$`},
}

func TestGenerate(t *testing.T) {
	for _, test := range cases {
		for i := 0; i < 10; i++ {
			res, err := GenerateFromString(test.regex, 10, initRand())
			if err != nil {
				t.Fatal("Error with regex: ", err)
			}
			// only print first result
			if i < 1 {
				fmt.Printf("Regex: %v Result: \"%s\"\n", test.regex, res)
			}
			re, err := regexp.Compile(test.regex)
			if err != nil {
				t.Fatal("Invalid test case. regex: ", test.regex, " failed to compile:", err)
			}
			if !re.MatchString(res) {
				t.Error("Generated data does not match regex. Regex: ", test.regex, " output: ", res)
			}
		}
	}
}

func TestSeed(t *testing.T) {
	currentTime := time.Now().UnixNano()
	rand1 := rand.New(rand.NewSource(currentTime))
	rand2 := rand.New(rand.NewSource(currentTime))
	for i := 0; i < 10; i++ {
		res1, err1 := GenerateFromString(cases[0].regex, 100, rand1)
		res2, err2 := GenerateFromString(cases[0].regex, 100, rand2)
		if err1 != nil || err2 != nil {
			t.Fatal("Error with regex: ", err1, err2)
		}
		if res1 != res2 {
			t.Error("Results are not reproducible")
		}
	}

	rand1 = rand.New(rand.NewSource(123))
	rand2 = rand.New(rand.NewSource(456))
	for i := 0; i < 10; i++ {
		res1, err1 := GenerateFromString(cases[0].regex, 100, rand1)
		res2, err2 := GenerateFromString(cases[0].regex, 100, rand2)
		if err1 != nil || err2 != nil {
			t.Fatal("Error with regex: ", err1, err2)
		}
		if res1 == res2 {
			t.Error("Results should not match")
		}
	}
}

func BenchmarkGenerate(b *testing.B) {
	regex, err := syntax.Parse(`^[a-z]{5,10}@[a-z]+\.(com|net|org)$`, syntax.Perl)
	if err != nil {
		b.Fatal("Error with regex ", err)
	}

	randTest := initRand()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Generate(regex, 10, randTest)
	}
}

func initRand() *rand.Rand {
	return rand.New(rand.NewSource(time.Now().UnixNano()))
}
