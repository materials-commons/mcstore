/*
Package config is a flexible tool kit that simplifies application configuration. It is modeled
after the standard library's io and net/http packages, and inspired by the inconshreveables'
log15 package. It allows you to easily set up logging the way you want for your application.
It provides implementations of other standard configuration methods, such as the approach
espoused for 12 factor apps (http://12factor.net/config), or the multi-level configuration
scheme in viper (https://github.com/spf13/viper).

Like the standard library net/http package, config relies heavily on standard interfaces,
in particular the Handler and Loader interfaces. These interfaces act like lego blocks
allowing you to combine them in customized ways.

Getting Started

To get started you will need to import the library:
   import "github.com/materials-commons/config"

You can now start to use the config package. The easiest way is to use
the standard config object. Unless you need to create multiple configuration
entries, its easiest to just use the standard, this is a package global
that you can access, for example config.Get(...).

Before you start using config you need to tell it how it will find its
configuration information. For example, lets say all your configuration
information is kept in environment variables. You can do:

import (
   "github.com/materials-commons/config"
   "github.com/materials-commons/config/handler"

func main() {
    config.Init(handler.Env())
}
*/
package config
