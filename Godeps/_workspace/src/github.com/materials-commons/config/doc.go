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
the standard config object. Before you start using config you need to tell
it how it how to find configuration information. For example, if your configuration
information is kept in environment variables you would setup the config as follows:

   import (
      "github.com/materials-commons/config"
      "github.com/materials-commons/config/handler"
   )

   func main() {
       config.Init(handler.Env())
   }

This will setup config to look for configuration entries by accessing
environment variables.

While you can setup your own config variable, for most projects you can use
the standard configuration. The standard configuration is accessible as package
global methods. When you load config, it will initialize a standard configuration
for you. It's default configuration is the TwelveFactor handler. You can easily
change the default handler. Lets take a look at how using this:

    import (
        "github.com/materials-commons/config"
        "github.com/materials-commons/config/handler"
    )

    func main() {
        config.Init(config.TwelveFactorWithOverride)
        port := config.GetInt("MYAPP_PORT")
    }

Handlers

The Handler interface allows you to define new handlers for your specific use case.
A Handler defines how config sets and gets configuration keys. Config comes with
a large array of Handlers. Some Handlers allow you to combine handlers together.
This is useful if you have a series of fallbacks to look for configuration keys.

For example, by default you might want your app to look for its configuration
in environment variables. If they can't be found there, you want it to look in
a YAML file. You could set this up by using the Multi Handler, and passing it
the Env Handler, and the YAML Handler. For good measure you want access to
be safe across multiple go routines. Here is what that looks like:

    import (
        "io/ioutil"
        "bytes"
        "github.com/materials-commons/config"
        "github.com/materials-commons/config/handler"
        "github.com/materials-commons/config/loader"
    )

    func main() {
         b, _ := ioutil.ReadFile("/etc/myapp/app.yaml")
	     l := loader.YAML(bytes.NewReader(b))
	     myAppHandler := handler.Sync(handler.Multi(handler.Env(), handler.Loader(l)))
         config.Init(myAppHandler)
         port := config.GetInt("MYAPP_PORT")
    }

*/
package config
