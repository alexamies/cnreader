# Chinese Text Reader

A command line app for Chinese text corpus index, translation memory, and HTML
reader page generation.

Many HTML pages at chinesenotes.com and family of sites are generated from Go
templates from the lexical database and text corpus. This readme gives
instructions. It assumes that you have already cloned the project from GitHub.

The app also compiles unigram and bigram index for full text searching of the
corpus.

In addition, a translation memory is compiled from the dictionary, named
entity database, and phrase memory.

## Setup

Install the Go SDK -
[Install Documentation](https://golang.org/doc/install)

Make sure that your the go executable is on your path. You may need to do 
something like 

```shell
$ export PATH=$PATH:/usr/local/go/bin
```

## Basic Use

### Get the source code and add the directory $CNREADER_HOME/go to your GOPATH

```shell
sudo apt-get install -y git
git clone https://github.com/alexamies/chinesenotes.com.git
cd chinesenotes.com
export CNREADER_HOME=`pwd`
cd go
```
### Build the project

```shell
cd $CNREADER_HOME/go/src/cnreader
go build
```
## Generate word definition files

```
./cnreader -hwfiles
```

## Analyze the whole, including word frequencies and writing out docs to HTML

```shell
cd $CNREADER_HOME/go/src/cnreader
./cnreader.go
```

### To enhance all files listed in data/corpus/html-conversions.csv

```shell
./cnreader -html
```

### To enhance all files in the corpus file modern_articles.csv

```shell
./cnreader -collection modern_articles.csv
```

### To build the headword file and add headword numbers to the words.txt file

```shell
cd $CNREADER_HOME
cp ../buddhist-dictionary/data/dictionary/words.txt data/.
cd $CNREADER_HOME/go/src/cndreader
./cnreader -headwords
cd ../util
go run headwords.go
cd $CNREADER_HOME
cp data/lexical_units.txt data/words.txt
cd ../cnreader
./cnreader -hwfiles
```

## Special cases

The character 著 is both a simplified character and a traditional character that
maps to the simplified character 着. It is not handled by the word detail
program at the moment. To fix it keep the entry:

```
971	著	\N	zhù	971, 16830, 41404
```
in the headwords.txt file. Some manual editing of the file words/971.html might
be needed.

### Run unit tests

```shell
cd $CNREADER_HOME/src/cnreader/analysis
go test
cd $CNREADER_HOME/src/cnreader/dictionary
go test
# Similarly for other packages
```

## Potential issues

If you run out of memory running the cnreader command then you may need to increase the locked memory. 
Edit the /etc/security/limits.conf file to increase this.

## Analyzing your own corpus

The cnreader program looks at the file $CNREADER_HOME/data/corpus/collections.csv and analyzes the lists of texts under there. To analyze your own corpus, create a new directory tree with your own collections.csv file and set the environment variable CNREADER_HOME to the top of that directory.

## Testing

Run unit tests with the command

```shell
go test ./... -cover
```

Run an integration test with the command

```shell
go test -integration ./...
```