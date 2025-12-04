package kingpin

import (
	"strconv"
	"strings"
	"time"
)

var flags = make(map[string]FlagReg)

type flagType int

const (
	flagTypeBool flagType = iota
	flagTypeString
	flagTypeInt
	flagTypeDuration
	flagTypeStrings
)

func Flag(name string, desc string) FlagReg {
	return FlagReg{
		Name:        name,
		Description: desc,
	}
}

type FlagReg struct {
	Name        string
	Description string
	DefVal      string
	Value       any
	Type        flagType
}

func (r FlagReg) Bool() *bool {
	value := new(bool)
	if strings.ToLower(r.DefVal) == "true" {
		t := true
		value = &t
	}
	flags[r.Name] = FlagReg{
		Name:        r.Name,
		Description: r.Description,
		DefVal:      r.DefVal,
		Value:       value,
		Type:        flagTypeBool,
	}
	return value
}

func (r FlagReg) String() *string {
	value := r.DefVal
	flags[r.Name] = FlagReg{
		Name:        r.Name,
		Description: r.Description,
		DefVal:      r.DefVal,
		Value:       &value,
		Type:        flagTypeString,
	}
	return &value
}

func (r FlagReg) Int() *int {
	value := new(int)
	if r.DefVal != "" {
		i, _ := strconv.Atoi(r.DefVal)
		value = &i
	}
	flags[r.Name] = FlagReg{
		Name:        r.Name,
		Description: r.Description,
		DefVal:      r.DefVal,
		Value:       value,
		Type:        flagTypeInt,
	}
	return value
}

func (r FlagReg) Duration() *time.Duration {
	value := new(time.Duration)
	flags[r.Name] = FlagReg{
		Name:        r.Name,
		Description: r.Description,
		DefVal:      r.DefVal,
		Value:       value,
		Type:        flagTypeDuration,
	}
	return value
}

func (r FlagReg) Strings() *[]string {
	value := new([]string)
	flags[r.Name] = FlagReg{
		Name:        r.Name,
		Description: r.Description,
		DefVal:      r.DefVal,
		Value:       value,
		Type:        flagTypeStrings,
	}
	return value
}

func (r FlagReg) Hidden() FlagReg {
	return r
}

func (r FlagReg) Default(value string) FlagReg {
	return FlagReg{
		Name:        r.Name,
		Description: r.Description,
		DefVal:      value,
		Value:       r.Value,
	}
}

type ParseContext struct {
}

func (r FlagReg) PreAction(f func(c *ParseContext) error) FlagReg {
	return r
}

func (r FlagReg) Action(action func(ctx *ParseContext) error) FlagReg {
	return r
}

func (r FlagReg) Envar(s string) FlagReg {
	return r
}

type CLI struct {
}

func (CLI) Parse(args []string) (string, error) {
	return "", nil
}

var (
	CommandLine = &CLI{}
)

func Set(key string, value string) {
	flag, found := flags[key]
	if !found {
		return
	}
	switch flag.Type {
	case flagTypeBool:
		v := strings.ToLower(value) == "true"
		*flag.Value.(*bool) = v
	case flagTypeString:
		*flag.Value.(*string) = value
	case flagTypeInt:
		i, _ := strconv.Atoi(value)
		*flag.Value.(*int) = i
	case flagTypeDuration:
		d, _ := time.ParseDuration(value)
		*flag.Value.(*time.Duration) = d
	case flagTypeStrings:
		split := strings.Split(value, ",")
		*flag.Value.(*[]string) = split
	}

}
