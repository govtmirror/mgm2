# mgm2

This project is a complete rewrite of the MOSES Grid Manager [MGM] application.  This application replaces the apache2/php backend with a Golang server; and replaces most ajax calls with websockets.  In order to function correctly, this application also relies on the MGMModule project, which is an Opensimulator plugin.

This project is generated with [yo angular generator](https://github.com/yeoman/generator-angular)
version 0.11.1.

## Vagrant

Vagrant is a cross-platform application for managing virtual machines.  It can be used to create a homogenous development environment that is easy to set up, is independent from any developers workstation packages (or platform), and matches the deployment environment.  For a quick overview, reference this blog post http://www.erikaheidi.com/blog/a-begginers-guide-to-vagrant-getting-your-portable-development-e

This application comes with a vagrant file and provinioning script for virtualbox using ubuntu 14.04 LTS.  If used, vagrant will create a virtualbox vm on your workstation and install the needed applications for development.

If you use vagrant, once `vagrant up` is complete, enter your vm using `vagrant ssh`.  The project root directory is synched to `/vagrant`, so `cd /vagrant` and run `go get`, `bower install`, `npm install` to download external libraries.  To run grunt for preview and front-end compilation, run `sudo npm install -g grunt-cli`.

## Environment setup

This application is comprised of a front-end and a back-end, both of which are compiled.

### front-end

Required packages:

* yeoman
* npm
* grunt
* bower

setup:  In the root folder of the project, execute `npm install` and `bower install` to download all libraries.

To add to the project, execute `yo angular:[type]` where type is as described at https://github.com/yeoman/generator-angular
yeoman will update seveal portions of the project, including the unit tests, so you can go directly to pop[ulating the newly created object.

### back-end

The backend is compiled go code.  Once you have Go installed on your system, and have $GOPATH and $GOBIN defined for your workspace, execute `go get` to download required libraries.

## Build & development

### front-end

Run `grunt` for building and `grunt serve` for preview.  Running `grunt test` will execute the testing suite without outputting to the dist folder.  The preview is a self-reloading on edit preview in a new tab of your default web browser.

### back-end

Run `go run mgmServer.go` to compile and execute from a system temporary directory.  Run `go install mgmServer.go` to compile the back-end and place the executable into your $GOBIN.

## Testing

### front-end
Running `grunt test` will run the unit tests with karma.

### back-end
Testing not yet in place.
