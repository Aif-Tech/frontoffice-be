# validator
[![GoDoc](https://godoc.org/github.com/usepzaka/validator?status.png)](https://godoc.org/github.com/usepzaka/validator)

===========

Validator Struct, Interface, Variable and Type for GO


#### Installation
Make sure that Go is installed on your computer.
Type the following command in your terminal:

	go get github.com/usepzaka/validator

After it the package is ready to use.


#### Import package in your project
Add following line in your `*.go` file:
```go
import "github.com/usepzaka/validator"
```
If you are unhappy to use long `validator`, you can do something like this:
```go
import (
  somethingname "github.com/usepzaka/validator"
)
```


#### List of functions:
```go
func Abs(value float64) float64
func BlackList(str, chars string) string
func ByteLength(str string, params ...string) bool
func CamelCaseToUnderscore(str string) string
func Contains(str, substring string) bool
func Count(array []interface{}, iterator ConditionIterator) int
func Each(array []interface{}, iterator Iterator)
func ErrorByField(e error, field string) string
func ErrorsByField(e error) map[string]string
func Filter(array []interface{}, iterator ConditionIterator) []interface{}
func Find(array []interface{}, iterator ConditionIterator) interface{}
func GetLine(s string, index int) (string, error)
func GetLines(s string) []string
func HasLowerCase(str string) bool
func HasUpperCase(str string) bool
func HasWhitespace(str string) bool
func HasWhitespaceOnly(str string) bool
func InRange(value interface{}, left interface{}, right interface{}) bool
func InRangeFloat32(value, left, right float32) bool
func InRangeFloat64(value, left, right float64) bool
func InRangeInt(value, left, right interface{}) bool
func IsASCII(str string) bool
func IsAlpha(str string) bool
func IsAlphaSpace(str string) bool
func IsAlphanumeric(str string) bool
func IsAlphanumericSpace(str string) bool
func IsPhone(str string) bool
func IsIndoMPhone(str string) bool
func IsNik(str string) bool
func IsBase64(str string) bool
func IsByteLength(str string, min, max int) bool
func IsCIDR(str string) bool
func IsCRC32(str string) bool
func IsCRC32b(str string) bool
func IsCreditCard(str string) bool
func IsDNSName(str string) bool
func IsDataURI(str string) bool
func IsDialString(str string) bool
func IsDivisibleBy(str, num string) bool
func IsEmail(str string) bool
func IsExistingEmail(email string) bool
func IsFilePath(str string) (bool, int)
func IsFloat(str string) bool
func IsFullWidth(str string) bool
func IsHalfWidth(str string) bool
func IsHash(str string, algorithm string) bool
func IsHexadecimal(str string) bool
func IsHexcolor(str string) bool
func IsHost(str string) bool
func IsIP(str string) bool
func IsIPv4(str string) bool
func IsIPv6(str string) bool
func IsISBN(str string, version int) bool
func IsISBN10(str string) bool
func IsISBN13(str string) bool
func IsISO3166Alpha2(str string) bool
func IsISO3166Alpha3(str string) bool
func IsISO4217(str string) bool
func IsISO693Alpha2(str string) bool
func IsISO693Alpha3b(str string) bool
func IsIn(str string, params ...string) bool
func IsInRaw(str string, params ...string) bool
func IsInt(str string) bool
func IsJSON(str string) bool
func IsLatitude(str string) bool
func IsLongitude(str string) bool
func IsLowerCase(str string) bool
func IsMAC(str string) bool
func IsMD4(str string) bool
func IsMD5(str string) bool
func IsMagnetURI(str string) bool
func IsMongoID(str string) bool
func IsMultibyte(str string) bool
func IsNatural(value float64) bool
func IsNegative(value float64) bool
func IsNonNegative(value float64) bool
func IsNonPositive(value float64) bool
func IsNotNull(str string) bool
func IsNull(str string) bool
func IsNumeric(str string) bool
func IsPort(str string) bool
func IsPositive(value float64) bool
func IsPrintableASCII(str string) bool
func IsRFC3339(str string) bool
func IsRFC3339WithoutZone(str string) bool
func IsRGBcolor(str string) bool
func IsRegex(str string) bool
func IsRequestURI(rawurl string) bool
func IsRequestURL(rawurl string) bool
func IsRipeMD128(str string) bool
func IsRipeMD160(str string) bool
func IsRsaPub(str string, params ...string) bool
func IsRsaPublicKey(str string, keylen int) bool
func IsSHA1(str string) bool
func IsSHA256(str string) bool
func IsSHA384(str string) bool
func IsSHA512(str string) bool
func IsSSN(str string) bool
func IsSemver(str string) bool
func IsTiger128(str string) bool
func IsTiger160(str string) bool
func IsTiger192(str string) bool
func IsTime(str string, format string) bool
func IsType(v interface{}, params ...string) bool
func IsURL(str string) bool
func IsUTFDigit(str string) bool
func IsUTFLetter(str string) bool
func IsUTFLetterNumeric(str string) bool
func IsUTFNumeric(str string) bool
func IsUUID(str string) bool
func IsUUIDv3(str string) bool
func IsUUIDv4(str string) bool
func IsUUIDv5(str string) bool
func IsULID(str string) bool
func IsUnixTime(str string) bool
func IsUpperCase(str string) bool
func IsVariableWidth(str string) bool
func IsWhole(value float64) bool
func LeftTrim(str, chars string) string
func Map(array []interface{}, iterator ResultIterator) []interface{}
func Matches(str, pattern string) bool
func MaxStringLength(str string, params ...string) bool
func MinStringLength(str string, params ...string) bool
func NormalizeEmail(str string) (string, error)
func PadBoth(str string, padStr string, padLen int) string
func PadLeft(str string, padStr string, padLen int) string
func PadRight(str string, padStr string, padLen int) string
func PrependPathToErrors(err error, path string) error
func Range(str string, params ...string) bool
func RemoveTags(s string) string
func ReplacePattern(str, pattern, replace string) string
func Reverse(s string) string
func RightTrim(str, chars string) string
func RuneLength(str string, params ...string) bool
func SafeFileName(str string) string
func SetFieldsRequiredByDefault(value bool)
func SetNilPtrAllowedByRequired(value bool)
func Sign(value float64) float64
func StringLength(str string, params ...string) bool
func StringMatches(s string, params ...string) bool
func StripLow(str string, keepNewLines bool) string
func ToBoolean(str string) (bool, error)
func ToFloat(str string) (float64, error)
func ToInt(value interface{}) (res int64, err error)
func ToJSON(obj interface{}) (string, error)
func ToString(obj interface{}) string
func Trim(str, chars string) string
func Truncate(str string, length int, ending string) string
func TruncatingErrorf(str string, args ...interface{}) error
func UnderscoreToCamelCase(s string) string
func ValidateMap(inputMap map[string]interface{}, validationMap map[string]interface{}) error
func ValidateStruct(s interface{}) error
func WhiteList(str, chars string) string
type ConditionIterator
type CustomTypeValidator
type Error
func (e Error) Error() string
type Errors
func (es Errors) Error() string
func (es Errors) Errors() []error
type ISO3166Entry
type ISO693Entry
type InterfaceParamValidator
type Iterator
type ParamValidator
type ResultIterator
type UnsupportedTypeError
func (e *UnsupportedTypeError) Error() string
type Validator
```

#### Examples
###### IsURL
```go
println(validator.IsURL(`http://user@pass:domain.com/path/page`))
```
###### IsType
```go
println(validator.IsType("Bob", "string"))
println(validator.IsType(1, "int"))
i := 1
println(validator.IsType(&i, "*int"))
```

IsType can be used through the tag `type` which is essential for map validation:
```go
type User	struct {
  Name string      `valid:"type(string)"`
  Age  int         `valid:"type(int)"`
  Meta interface{} `valid:"type(string)"`
}
err := validator.ValidateStruct(User{"Bob", 20, "meta"})
if err != nil {
	println("error: " + err.Error())
}

```
###### ToString
```go
type User struct {
	FirstName string
	LastName string
}

str := validator.ToString(&User{"John", "Juan"})
println(str)
```
###### Each, Map, Filter, Count for slices
Each iterates over the slice/array and calls Iterator for every item
```go
data := []interface{}{1, 2, 3, 4, 5}
var fn validator.Iterator = func(value interface{}, index int) {
	println(value.(int))
}
validator.Each(data, fn)
```
```go
data := []interface{}{1, 2, 3, 4, 5}
var fn validator.ResultIterator = func(value interface{}, index int) interface{} {
	return value.(int) * 3
}
_ = validator.Map(data, fn) // result = []interface{}{1, 6, 9, 12, 15}
```

```go
data := []interface{}{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
var fn validator.ConditionIterator = func(value interface{}, index int) bool {
	return value.(int)%2 == 0
}
_ = validator.Filter(data, fn) // result = []interface{}{2, 4, 6, 8, 10}
_ = validator.Count(data, fn) // result = 5
```
###### ValidateStruct
If you want to validate structs, you can use tag `validate` for any field in your structure. All validators used with this field in one tag are separated by comma. If you want to skip validation, place `-` in your tag. 

For completely custom validators (interface-based), see below.

Here is a list of available validators for struct fields (validator - used function):


| Tag | Description |
| - | - |
| "email" |               Email format value allowed|
| "url" |                URL format value allowed |
| "dialstring" |           Dial String format value allowed |
| "requrl"  |            RequestURL format value allowed |
| "requri"  |            RequestURI format value allowed |
| "alpha"   |            Alphabet format value allowed |
| "alphaspace" |        Alphabet and whiteSpace allowee |
| "utfletter" |        UTF Letter format value allowed |
| "alphanum" |          Alphabet and numeric format value allowed |
| "alphanumspace" |      Alphabet numeric and whitespace format value allowed|
| "word" |               Alphabet, numeric, (-) and (_)values are allowed |
| "text" |          Alphabet, numeric, -, _, +, ;, &, /, \,  (), : and whitespace values are allowed |
| "phone" |			  Phone format without space are allowed |
| "indophone" |		  Indonesia phone number format are allowed. (Ex: 628xxxx. 08xxxx, +628xxxxx) |
| "utfletternum" |       UTF Letter values are allowed |
| "numeric" |            Numeric values are allowed |
| "utfnumeric" |         UTF Numeric values are allowed |
| "utfdigit" |           UTF Digit values are allowed |
| "hexadecimal" |        Hexadecmal values are allowed |
| "hexcolor" |          Hex COlor values are allowed |
| "rgbcolor" |            RGB color  values are allowed |
| "lowercase" |          Lower-Case values are allowed  |
| "uppercase" |         Upper-Case  values are allowed  |
| "int" |         Int values are allowed  |
| "float" |              Float values are allowed  |
| "null" |               Null values are allowed  |
| "uuid" |               UUID values are allowed  |
| "uuidv3" |             UUIDv3 values are allowed  |
| "uuidv4" |             UUID values are allowed  |
| "uuidv5" |             UUID values are allowed  |
| "creditcard" |         Credit Card values are allowed  |
| "isbn10" |            ISBN10 values are allowed  |
| "isbn13" |             ISBN13 values are allowed  |
| "json" |               JSON values are allowed  |
| "multibyte" |          Multibyte values are allowed  |
| "ascii" |              ASCII values are allowed  |
| "printableascii" |     PrintableASCII values are allowed  |
| "fullwidth" |         FullWidth values are allowed  |
| "halfwidth" |          HalfWidth values are allowed  |
| "variablewidth" |      VariableWidth values are allowed  |
| "base64" |             Base64 values are allowed  |
| "datauri" |            DataURI values are allowed  |
| "ip" |                 IP values are allowed  |
| "port" |               Port values are allowed  |
| "ipv4" |              IPv4 values are allowed  |
| "ipv6" |              IPv6 | values are allowed 
| "dns" |              DNSName values are allowed  |
| "host" |               Host values are allowed  |
| "mac" |                MAC values are allowed  |
| "latitude" |           Latitude values are allowed  |
| "longitude" |          Longitude values are allowed  |
| "ssn" |                SSN values are allowed  |
| "semver" |             Semver values are allowed  |
| "rfc3339" |            RFC3339 values are allowed  |
| "rfc3339WithoutZone" | RFC3339WithoutZone values are allowed  |
| "ISO3166Alpha2" |      ISO3166Alpha2 values are allowed  |
| "ISO3166Alpha3" |     ISO3166Alpha3 values are allowed  |
| "ulid" |               ULID values are allowed  |

Validators with parameters

| Tag | Use Function |
| - | - |
| `length(int)` | validate length of string |
| `range(min\|max)` | Range  |
| `min(int)` | Minimum String length  |
| `max(int)` | Maximum String length |
| `bytelength(min\|max)` |  Length of byte |
| `runelength(min\|max)` | Length of Rune |
| `matches(pattern)` | String matches of regular expression pattern |
| `in(string1\|string2\|...\|stringN)` | String is a member of the set of strings params |
| `rsapub(keylength)` | String is a RSApub |


Validators with parameters for any type

| Tag | Use Function |
| - | - |
| `type(type)` | validate type of struct|


And here is small example of usage:
```go
type Post struct {
	Title    string `valid:"alphanum,required"`
	Message2 string `valid:"type(string)"`
	IpAddress string `valid:"ipv4"`
	Phone     string `valid:"phone"`
}
post := &Post{
	Title:   "My Example Post",
	Message: "duck",
	IpAddress: "123.234.54.3",
	Message2: "087812345678",
}
```
###### ValidateMap 
If you want to validate maps, you can use the map to be validated and a validation map that contain the same tags used in ValidateStruct, both maps have to be in the form `map[string]interface{}`

So here is small example of usage:
```go
var mapTemplate = map[string]interface{}{
	"name":"required,alphaspace",
	"family":"required,alphaspace",
	"email":"required,email",
	"cell-phone":"phone",
	"address":map[string]interface{}{
		"line1":"required,alphanumspace",
		"line2":"text",
		"postal-code":"numeric",
	},
}

var inputMap = map[string]interface{}{
	"name":"Bob",
	"family":"Smith",
	"email":"foo@bar.com",
	"address":map[string]interface{}{
		"line1":"",
		"line2":"",
		"postal-code":"",
	},
}

err := validator.ValidateMap(inputMap, mapTemplate)
if err != nil {
	println("error: " + err.Error())
}

```

###### WhiteList
```go
// Remove all characters from string ignoring characters between "a" and "z"
println(validator.WhiteList("a3a43a5a4a3a2a23a4a5a4a3a4", "a-z") == "aaaaaaaaaaaa")
```

###### Loop over Error()
By default .Error() returns all errors in a single String. To access each error you can do this:
```go
  if err != nil {
    errs := err.(validator.Errors).Errors()
    for _, e := range errs {
      fmt.Println(e.Error())
    }
  }
```

###### Custom error messages
Custom error messages are supported via annotations by adding the `~` separator - here's an example of how to use it:
```go
type Ticket struct {
  Id        int64     `json:"id"`
  FirstName string    `json:"firstname" valid:"required~First name is blank"`
}
```

