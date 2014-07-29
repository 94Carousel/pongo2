package pongo2

/* Missing filters:

   safeseq
   truncatechars_html
   truncatewords_html

   Filters that won't be added:

   get_static_prefix (reason: web-framework specific)
   pprint (reason: python-specific)
   static (reason: web-framework specific)

   Rethink:

   force_escape (reason: not yet needed since this is the behaviour of pongo2's escape filter)
   unordered_list (python-specific; not sure whether needed or not)
   dictsort (python-specific; maybe one could add a filter to sort a list of structs by a specific field name)
   dictsortreversed (see dictsort)

   Filters that are provided through github.com/flosch/pongo2-addons:

   filesizeformat
   slugify
   timesince
   timeuntil
*/

import (
	"bytes"
	"errors"
	"fmt"
	"math/rand"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"
)

func init() {
	rand.Seed(time.Now().Unix())

	RegisterFilter("escape", filterEscape)
	RegisterFilter("safe", filterSafe)
	RegisterFilter("escapejs", filterEscapejs)

	RegisterFilter("add", filterAdd)
	RegisterFilter("addslashes", filterAddslashes)
	RegisterFilter("capfirst", filterCapfirst)
	RegisterFilter("center", filterCenter)
	RegisterFilter("cut", filterCut)
	RegisterFilter("date", filterDate)
	RegisterFilter("default", filterDefault)
	RegisterFilter("default_if_none", filterDefaultIfNone)
	RegisterFilter("divisibleby", filterDivisibleby)
	RegisterFilter("first", filterFirst)
	RegisterFilter("floatformat", filterFloatformat)
	RegisterFilter("get_digit", filterGetdigit)
	RegisterFilter("iriencode", filterIriencode)
	RegisterFilter("join", filterJoin)
	RegisterFilter("last", filterLast)
	RegisterFilter("length", filterLength)
	RegisterFilter("length_is", filterLengthis)
	RegisterFilter("linebreaks", filterLinebreaks)
	RegisterFilter("linebreaksbr", filterLinebreaksbr)
	RegisterFilter("linenumbers", filterLinenumbers)
	RegisterFilter("ljust", filterLjust)
	RegisterFilter("lower", filterLower)
	RegisterFilter("make_list", filterMakelist)
	RegisterFilter("phone2numeric", filterPhone2numeric)
	RegisterFilter("pluralize", filterPluralize)
	RegisterFilter("random", filterRandom)
	RegisterFilter("removetags", filterRemovetags)
	RegisterFilter("rjust", filterRjust)
	RegisterFilter("slice", filterSlice)
	RegisterFilter("stringformat", filterStringformat)
	RegisterFilter("striptags", filterStriptags)
	RegisterFilter("time", filterDate) // time uses filterDate (same golang-format)
	RegisterFilter("title", filterTitle)
	RegisterFilter("truncatechars", filterTruncatechars)
	RegisterFilter("truncatewords", filterTruncatewords)
	RegisterFilter("upper", filterUpper)
	RegisterFilter("urlencode", filterUrlencode)
	RegisterFilter("urlize", filterUrlize)
	RegisterFilter("urlizetrunc", filterUrlizetrunc)
	RegisterFilter("wordcount", filterWordcount)
	RegisterFilter("wordwrap", filterWordwrap)
	RegisterFilter("yesno", filterYesno)

	RegisterFilter("float", filterFloat)     // pongo-specific
	RegisterFilter("integer", filterInteger) // pongo-specific
}

func filterTruncatechars(in *Value, param *Value) (*Value, error) {
	s := in.String()
	newLen := param.Integer()
	if newLen < len(s) {
		if newLen >= 3 {
			return AsValue(fmt.Sprintf("%s...", s[:newLen-3])), nil
		}
		// Not enough space for the ellipsis
		return AsValue(s[:newLen]), nil
	}
	return in, nil
}

func filterTruncatewords(in *Value, param *Value) (*Value, error) {
	words := strings.Fields(in.String())
	n := param.Integer()
	if n <= 0 {
		return AsValue(""), nil
	}
	nlen := min(len(words), n)
	out := make([]string, 0, nlen)
	for i := 0; i < nlen; i++ {
		out = append(out, words[i])
	}

	if n < len(words) {
		out = append(out, "...")
	}

	return AsValue(strings.Join(out, " ")), nil
}

func filterEscape(in *Value, param *Value) (*Value, error) {
	output := strings.Replace(in.String(), "&", "&amp;", -1)
	output = strings.Replace(output, ">", "&gt;", -1)
	output = strings.Replace(output, "<", "&lt;", -1)
	output = strings.Replace(output, "\"", "&quot;", -1)
	output = strings.Replace(output, "'", "&#39;", -1)
	return AsValue(output), nil
}

func filterSafe(in *Value, param *Value) (*Value, error) {
	return in, nil // nothing to do here, just to keep track of the safe application
}

func filterEscapejs(in *Value, param *Value) (*Value, error) {
	sin := in.String()
	l := len(sin)

	var b bytes.Buffer

	idx := 0
	for idx < l {
		c := rune(sin[idx])

		if c == '\\' {
			// Escape seq?
			if idx + 1 < l {
				switch sin[idx+1] {
				case 'r':
					b.WriteString(fmt.Sprintf(`\\u%04X`, '\r'))
					idx += 2
					continue
				case 'n':
					b.WriteString(fmt.Sprintf(`\\u%04X`, '\n'))
					idx += 2
					continue
				case '\'':
					b.WriteString(fmt.Sprintf(`\\u%04X`, '\''))
					idx += 2
					continue
				}
			}
		}

		if (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || c == ' ' || c == '/' {
			b.WriteRune(c)
		} else {
			b.WriteString(fmt.Sprintf(`\\u%04X`, c))
		}
		idx++
	}

	return AsValue(b.String()), nil
}

func filterAdd(in *Value, param *Value) (*Value, error) {
	if in.IsNumber() && param.IsNumber() {
		if in.IsFloat() || param.IsFloat() {
			return AsValue(in.Float() + param.Float()), nil
		} else {
			return AsValue(in.Integer() + param.Integer()), nil
		}
	}
	// If in/param is not a number, we're relying on the
	// Value's String() convertion and just add them both together
	return AsValue(in.String() + param.String()), nil
}

func filterAddslashes(in *Value, param *Value) (*Value, error) {
	output := strings.Replace(in.String(), "\\", "\\\\", -1)
	output = strings.Replace(output, "\"", "\\\"", -1)
	output = strings.Replace(output, "'", "\\'", -1)
	return AsValue(output), nil
}

func filterCut(in *Value, param *Value) (*Value, error) {
	return AsValue(strings.Replace(in.String(), param.String(), "", -1)), nil
}

func filterLength(in *Value, param *Value) (*Value, error) {
	return AsValue(in.Len()), nil
}

func filterLengthis(in *Value, param *Value) (*Value, error) {
	return AsValue(in.Len() == param.Integer()), nil
}

func filterDefault(in *Value, param *Value) (*Value, error) {
	if !in.IsTrue() {
		return param, nil
	}
	return in, nil
}

func filterDefaultIfNone(in *Value, param *Value) (*Value, error) {
	if in.IsNil() {
		return param, nil
	}
	return in, nil
}

func filterDivisibleby(in *Value, param *Value) (*Value, error) {
	if param.Integer() == 0 {
		return AsValue(false), nil
	}
	return AsValue(in.Integer()%param.Integer() == 0), nil
}

func filterFirst(in *Value, param *Value) (*Value, error) {
	if in.CanSlice() && in.Len() > 0 {
		return in.Index(0), nil
	}
	return AsValue(""), nil
}

func filterFloatformat(in *Value, param *Value) (*Value, error) {
	val := in.Float()

	decimals := -1
	if !param.IsNil() {
		// Any argument provided?
		decimals = param.Integer()
	}

	// if the argument is not a number (e. g. empty), the default
	// behaviour is trim the result
	trim := !param.IsNumber()

	if decimals <= 0 {
		// argument is negative or zero, so we
		// want the output being trimmed
		decimals = -decimals
		trim = true
	}

	if trim {
		// Remove zeroes
		if float64(int(val)) == val {
			return AsValue(in.Integer()), nil
		}
	}

	return AsValue(strconv.FormatFloat(val, 'f', decimals, 64)), nil
}

func filterGetdigit(in *Value, param *Value) (*Value, error) {
	i := param.Integer()
	l := len(in.String()) // do NOT use in.Len() here!
	if i <= 0 || i > l {
		return in, nil
	}
	return AsValue(in.String()[l-i] - 48), nil
}

const filterIRIChars = "/#%[]=:;$&()+,!?*@'~"

func filterIriencode(in *Value, param *Value) (*Value, error) {
	var b bytes.Buffer

	sin := in.String()
	for _, r := range sin {
		if strings.IndexRune(filterIRIChars, r) >= 0 {
			b.WriteRune(r)
		} else {
			b.WriteString(url.QueryEscape(string(r)))
		}
	}

	return AsValue(b.String()), nil
}

func filterJoin(in *Value, param *Value) (*Value, error) {
	if !in.CanSlice() {
		return in, nil
	}
	sep := param.String()
	sl := make([]string, 0, in.Len())
	for i := 0; i < in.Len(); i++ {
		sl = append(sl, in.Index(i).String())
	}
	return AsValue(strings.Join(sl, sep)), nil
}

func filterLast(in *Value, param *Value) (*Value, error) {
	if in.CanSlice() && in.Len() > 0 {
		return in.Index(in.Len() - 1), nil
	}
	return AsValue(""), nil
}

func filterUpper(in *Value, param *Value) (*Value, error) {
	return AsValue(strings.ToUpper(in.String())), nil
}

func filterLower(in *Value, param *Value) (*Value, error) {
	return AsValue(strings.ToLower(in.String())), nil
}

func filterMakelist(in *Value, param *Value) (*Value, error) {
	s := in.String()
	result := make([]string, 0, len(s))
	for _, c := range s {
		result = append(result, string(c))
	}
	return AsValue(result), nil
}

func filterCapfirst(in *Value, param *Value) (*Value, error) {
	if in.Len() <= 0 {
		return AsValue(""), nil
	}
	t := in.String()
	return AsValue(strings.ToUpper(string(t[0])) + t[1:]), nil
}

func filterCenter(in *Value, param *Value) (*Value, error) {
	width := param.Integer()
	slen := in.Len()
	if width <= slen {
		return in, nil
	}

	spaces := width - slen
	left := spaces/2 + spaces%2
	right := spaces / 2

	return AsValue(fmt.Sprintf("%s%s%s", strings.Repeat(" ", left),
		in.String(), strings.Repeat(" ", right))), nil
}

func filterDate(in *Value, param *Value) (*Value, error) {
	t, is_time := in.Interface().(time.Time)
	if !is_time {
		return nil, errors.New("Filter input argument must be of type 'time.Time'.")
	}
	return AsValue(t.Format(param.String())), nil
}

func filterFloat(in *Value, param *Value) (*Value, error) {
	return AsValue(in.Float()), nil
}

func filterInteger(in *Value, param *Value) (*Value, error) {
	return AsValue(in.Integer()), nil
}

func filterLinebreaks(in *Value, param *Value) (*Value, error) {
	if in.Len() == 0 {
		return in, nil
	}

	var b bytes.Buffer

	// Newline = <br />
	// Double newline = <p>...</p>
	lines := strings.Split(in.String(), "\n")
	lenlines := len(lines)

	opened := false

	for idx, line := range lines {

		if !opened {
			b.WriteString("<p>")
			opened = true
		}

		b.WriteString(line)

		if idx < lenlines-1 && strings.TrimSpace(lines[idx]) != "" {
			// We've not reached the end
			if strings.TrimSpace(lines[idx+1]) == "" {
				// Next line is empty
				if opened {
					b.WriteString("</p>")
					opened = false
				}
			} else {
				b.WriteString("<br />")
			}
		}
	}

	if opened {
		b.WriteString("</p>")
	}

	return AsValue(b.String()), nil
}

func filterLinebreaksbr(in *Value, param *Value) (*Value, error) {
	return AsValue(strings.Replace(in.String(), "\n", "<br />", -1)), nil
}

func filterLinenumbers(in *Value, param *Value) (*Value, error) {
	lines := strings.Split(in.String(), "\n")
	output := make([]string, 0, len(lines))
	for idx, line := range lines {
		output = append(output, fmt.Sprintf("%d. %s", idx+1, line))
	}
	return AsValue(strings.Join(output, "\n")), nil
}

func filterLjust(in *Value, param *Value) (*Value, error) {
	times := param.Integer() - in.Len()
	if times < 0 {
		times = 0
	}
	return AsValue(fmt.Sprintf("%s%s", in.String(), strings.Repeat(" ", times))), nil
}

func filterUrlencode(in *Value, param *Value) (*Value, error) {
	return AsValue(url.QueryEscape(in.String())), nil
}

// TODO: This regexp could do some work
var filterUrlizeURLRegexp = regexp.MustCompile(`((((http|https)://)|www\.|\w+(\.com|\.net|\.org|\.info|\.biz|\.de)/))(?U:.*)[ ]`)
var filterUrlizeEmailRegexp = regexp.MustCompile(`(\w+@\w+\.\w{2,4})`)

func filterUrlizeHelper(input string, autoescape bool, trunc int) string {
	sout := filterUrlizeURLRegexp.ReplaceAllStringFunc(input, func(raw_url string) string {
		raw_url = strings.TrimSpace(raw_url)

		t, err := ApplyFilter("iriencode", AsValue(raw_url), nil)
		if err != nil {
			panic(err)
		}
		url := t.String()

		if !strings.HasPrefix(url, "http") {
			url = fmt.Sprintf("http://%s", url)
		}

		title := raw_url

		if trunc > 3 && len(title) > trunc {
			title = fmt.Sprintf("%s...", title[:trunc-3])
		}

		if autoescape {
			t, err := ApplyFilter("escape", AsValue(title), nil)
			if err != nil {
				panic(err)
			}
			title = t.String()
		}

		return fmt.Sprintf(`<a href="%s" rel="nofollow">%s</a> `, url, title)
	})

	sout = filterUrlizeEmailRegexp.ReplaceAllStringFunc(sout, func(mail string) string {

		title := mail

		if trunc > 3 && len(title) > trunc {
			title = fmt.Sprintf("%s...", title[:trunc-3])
		}

		return fmt.Sprintf(`<a href="mailto:%s">%s</a>`, mail, title)
	})

	return sout
}

func filterUrlize(in *Value, param *Value) (*Value, error) {
	autoescape := true
	if param.IsBool() {
		autoescape = param.Bool()
	}

	return AsValue(filterUrlizeHelper(in.String(), autoescape, -1)), nil
}

func filterUrlizetrunc(in *Value, param *Value) (*Value, error) {
	return AsValue(filterUrlizeHelper(in.String(), true, param.Integer())), nil
}

func filterStringformat(in *Value, param *Value) (*Value, error) {
	return AsValue(fmt.Sprintf(param.String(), in.Interface())), nil
}

var re_striptags = regexp.MustCompile("<[^>]*?>")

func filterStriptags(in *Value, param *Value) (*Value, error) {
	s := in.String()

	// Strip all tags
	s = re_striptags.ReplaceAllString(s, "")

	return AsValue(strings.TrimSpace(s)), nil
}

// https://en.wikipedia.org/wiki/Phoneword
var filterPhone2numericMap = map[string]string{
	"a": "2", "b": "2", "c": "2", "d": "3", "e": "3", "f": "3", "g": "4", "h": "4", "i": "4", "j": "5", "k": "5",
	"l": "5", "m": "6", "n": "6", "o": "6", "p": "7", "q": "7", "r": "7", "s": "7", "t": "8", "u": "8", "v": "8",
	"w": "9", "x": "9", "y": "9", "z": "9",
}

func filterPhone2numeric(in *Value, param *Value) (*Value, error) {
	sin := in.String()
	for k, v := range filterPhone2numericMap {
		sin = strings.Replace(sin, k, v, -1)
		sin = strings.Replace(sin, strings.ToUpper(k), v, -1)
	}
	return AsValue(sin), nil
}

func filterPluralize(in *Value, param *Value) (*Value, error) {
	if in.IsNumber() {
		// Works only on numbers
		if param.Len() > 0 {
			endings := strings.Split(param.String(), ",")
			if len(endings) > 2 {
				return nil, errors.New("You cannot pass more than 2 arguments to filter 'pluralize'.")
			}
			if len(endings) == 1 {
				// 1 argument
				if in.Integer() != 1 {
					return AsValue(endings[0]), nil
				}
			} else {
				if in.Integer() != 1 {
					// 2 arguments
					return AsValue(endings[1]), nil
				}
				return AsValue(endings[0]), nil
			}
		} else {
			if in.Integer() != 1 {
				// return default 's'
				return AsValue("s"), nil
			}
		}

		return AsValue(""), nil
	} else {
		return nil, errors.New("Filter 'pluralize' does only work on numbers.")
	}
}

func filterRandom(in *Value, param *Value) (*Value, error) {
	if !in.CanSlice() || in.Len() <= 0 {
		return in, nil
	}
	i := rand.Intn(in.Len())
	return in.Index(i), nil
}

func filterRemovetags(in *Value, param *Value) (*Value, error) {
	s := in.String()
	tags := strings.Split(param.String(), ",")

	// Strip only specific tags
	for _, tag := range tags {
		re := regexp.MustCompile(fmt.Sprintf("</?%s/?>", tag))
		s = re.ReplaceAllString(s, "")
	}

	return AsValue(strings.TrimSpace(s)), nil
}

func filterRjust(in *Value, param *Value) (*Value, error) {
	return AsValue(fmt.Sprintf(fmt.Sprintf("%%%ds", param.Integer()), in.String())), nil
}

func filterSlice(in *Value, param *Value) (*Value, error) {
	comp := strings.Split(param.String(), ":")
	if len(comp) != 2 {
		return nil, errors.New("Slice string must have the format 'from:to' [from/to can be omitted, but the ':' is required]")
	}

	if !in.CanSlice() {
		return in, nil
	}

	from := AsValue(comp[0]).Integer()
	to := in.Len()

	if from > to {
		from = to
	}

	vto := AsValue(comp[1]).Integer()
	if vto >= from && vto <= in.Len() {
		to = vto
	}

	return in.Slice(from, to), nil
}

func filterTitle(in *Value, param *Value) (*Value, error) {
	if !in.IsString() {
		return AsValue(""), nil
	}
	return AsValue(strings.Title(strings.ToLower(in.String()))), nil
}

func filterWordcount(in *Value, param *Value) (*Value, error) {
	return AsValue(len(strings.Fields(in.String()))), nil
}

func filterWordwrap(in *Value, param *Value) (*Value, error) {
	words := strings.Fields(in.String())
	words_len := len(words)
	wrap_at := param.Integer()
	if wrap_at <= 0 {
		return in, nil
	}

	linecount := words_len/wrap_at + words_len%wrap_at
	lines := make([]string, 0, linecount)
	for i := 0; i < linecount; i++ {
		lines = append(lines, strings.Join(words[wrap_at*i:min(wrap_at*(i+1), words_len)], " "))
	}
	return AsValue(strings.Join(lines, "\n")), nil
}

func filterYesno(in *Value, param *Value) (*Value, error) {
	choices := map[int]string{
		0: "yes",
		1: "no",
		2: "maybe",
	}
	param_string := param.String()
	custom_choices := strings.Split(param_string, ",")
	if len(param_string) > 0 {
		if len(custom_choices) > 3 {
			return nil, errors.New(fmt.Sprintf("You cannot pass more than 3 options to the 'yesno'-filter (got: '%s').", param_string))
		}
		if len(custom_choices) < 2 {
			return nil, errors.New(fmt.Sprintf("You must pass either no or at least 2 arguments to the 'yesno'-filter (got: '%s').", param_string))
		}

		// Map to the options now
		choices[0] = custom_choices[0]
		choices[1] = custom_choices[1]
		if len(custom_choices) == 3 {
			choices[2] = custom_choices[2]
		}
	}

	// maybe
	if in.IsNil() {
		return AsValue(choices[2]), nil
	}

	// yes
	if in.IsTrue() {
		return AsValue(choices[0]), nil
	}

	// no
	return AsValue(choices[1]), nil
}
