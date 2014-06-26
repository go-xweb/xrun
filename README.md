Xrun is a development tool for Go programming language.

===

# Features

* Auto rebuild and run the program when go files or other special files changed.

# Installation

If you have [got](http://github.com/gobuild/got) installed,

	got go-xweb/xrun

or

	go get github.com/go-xweb/xrun

# How to use

	cd <project dir>
	xrun

or

	cd <project dir>
	xrun main.go

# Add special files need to be monitered

Defaultly, xrun will moniter all the .go files. If you want to moniter other files, you can add a config file in the project root folder named xrun.json. The content just like below:

	{
		"Mode":1,
		"ExcludeDirs": {
			".git":true,
			".svn":true
		},
		"ExcludeFiles": {
		},
		"IncludeFiles": {
		},
		"IncludeDirs": {
		}
	}

Above is the default config. 
The `Mode` is the log level, default is `Linfo`:

	Ldebug = iota
	Linfo
	Lwarn
	Lerror
	Lpanic
	Lfatal
	Lnone