xRun is a development tool for Go programming language.

===

# Features

* Auto rebuild and run the program when go files or other special files be changed.

# How to use

	cd <project dir>
	xrun

or

	cd <project dir>
	xrun main.go

# Add special files need to be monitered

Defaultly, xrun will moniter all the .go files. If you want to moniter other files, you can add a config file in the project root folder named xrun.json. The content just like below:

	{
		"excludeDirs": {
			".git":true,
			".svn":true
		},
		"excludeFiles": {
		},
		"includeFiles": {
		}
	}

Above is the default config.