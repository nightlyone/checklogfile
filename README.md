checklogfile
===========

NOTE: Work in progress at the moment.

[![Build Status][1]][2]

[1]: https://secure.travis-ci.org/nightlyone/checklogfile.png
[2]: http://travis-ci.org/nightlyone/checklogfile


LICENSE
-------
BSD

documentation
-------------
[package documentation at go.pkgdoc.org](http://go.pkgdoc.org/github.com/nightlyone/checklogfile)


quick usage
-----------

TODO

build and install
=================

install from source
-------------------

Install [Go 1][3], either [from source][4] or [with a prepackaged binary][5].

Then run

	go get github.com/nightlyone/checklogfile
	go get github.com/nightlyone/checklogfile/cmd/check_logfile

Usage via

	$GOPATH/bin/check_logfile -h


[3]: http://golang.org
[4]: http://golang.org/doc/install/source
[5]: http://golang.org/doc/install

LICENSE
-------
BSD

documentation
=============
TODO
contributing
============

Contributions are welcome. Please open an issue or send me a pull request for a dedicated branch.
Make sure the git commit hooks show it works.

git commit hooks
-----------------------
enable commit hooks via

        cd .git ; rm -rf hooks; ln -s ../git-hooks hooks ; cd ..

