package configuration

import "errors"

var ErrorAliasExists = errors.New("alias already exists")
var ErrorAliasDontExists = errors.New("alias doesn't exists")
