// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// 
// Command line utility to analyze Chinese text, including corpus analysis, 
// compilation of a full text search index, and mark up HTML files in reader
// style.
//
// Quickstart
//
// Supply Chinese text on the command line. Observe tokenization and matching to
// English equivalents
//
// go get github.com/alexamies/cnreader
// go run github.com/alexamies/cnreader -source_text="君不見黃河之水天上來"
//
package main

import (
	"bufio"
	"encoding/csv"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"runtime/pprof"
	"strconv"
	"text/template"
	"time"

	"github.com/alexamies/chinesenotes-go/config"
	"github.com/alexamies/chinesenotes-go/dictionary"
	"github.com/alexamies/chinesenotes-go/dicttypes"	
	"github.com/alexamies/chinesenotes-go/fileloader"
	"github.com/alexamies/chinesenotes-go/tokenizer"
	"github.com/alexamies/cnreader/analysis"
	"github.com/alexamies/cnreader/corpus"
	"github.com/alexamies/cnreader/generator"
	"github.com/alexamies/cnreader/index"
	"github.com/alexamies/cnreader/library"
	"github.com/alexamies/cnreader/tmindex"
)

const conversionsFile = "data/corpus/html-conversion.csv"

// A type that holds the source and destination files for HTML conversion
type htmlConversion struct {
	SrcFile, DestFile, Template string
	GlossChinese bool
	Title string
}

// Initialize the app config data
func initApp() (config.AppConfig) {
	return config.InitConfig()
}

// formatTokens formats text tokens as plain text
func formatTokens(w io.Writer, tokens []tokenizer.TextToken) {
	for _, t := range tokens {
		fmt.Fprintf(w, "Token: %s\n", t.Token)
		fmt.Fprintf(w, "Pinyin: %s\n", t.DictEntry.Pinyin)
		if len(t.DictEntry.Senses) == 1 {
			fmt.Fprintf(w, "English: %s\n\n", t.DictEntry.Senses[0].English)
			continue
		}
		fmt.Fprintf(w, "English:\n")
		for i, ws := range t.DictEntry.Senses {
			j := i+1
			fmt.Fprintf(w, "%d. %s\n", j, ws.English)
		}
		fmt.Fprintln(w)
	} 
}

// getDocFreq gets the word document frequency
//
// If it cannot be read from file, it will be computed from the corpus
func getDocFreq(c config.AppConfig) (*index.DocumentFrequency, error) {
	fname := c.ProjectHome + "/" + library.LibraryFile
	corpusConfig := getCorpusConfig(c)
	libraryLoader := library.NewLibraryLoader(fname, corpusConfig)
	indexConfig := getIndexConfig(c)
	wdict, err := fileloader.LoadDictFile(c)
	if err != nil {
		return nil, fmt.Errorf("getDocFreq, error opening dictionary: %v", err)
	}
	dictTokenizer := tokenizer.DictTokenizer{wdict}
	dir := indexConfig.IndexDir
	dfFile, err := os.Open(dir + "/" + index.DocFreqFile)
	if err != nil {
		log.Printf("getDocFreq, error opening word freq file (recoverable): %v", err)
		df, err := analysis.GetDocFrequencies(libraryLoader, dictTokenizer,
				wdict)
		if err != nil {
			return nil, fmt.Errorf("getDocFreq: error computing document freq: %v", err)
		}
		return df, nil
	}
	defer dfFile.Close()
	df, err := index.ReadDocumentFrequency(dfFile)
	if err != nil {
		log.Printf("getDocFreq, error reading document frequency (recoverable): %v",
			err)
		df, err = analysis.GetDocFrequencies(libraryLoader, dictTokenizer,
				wdict)
		if err != nil {
			return nil, fmt.Errorf("getDocFreq, error computing doc freq: %v", err)
		}
	}
	return df, nil
}

// Gets the list of source and destination files for HTML conversion
func getHTMLConversions(c config.AppConfig) []htmlConversion {
	log.Printf("GetHTMLConversions: projectHome: '%s'\n", c.ProjectHome)
	conversionsFile := c.ProjectHome + "/" + conversionsFile
	convFile, err := os.Open(conversionsFile)
	if err != nil {
		log.Fatalf("getHTMLConversions.GetHTMLConversions fatal error: %v", err)
	}
	defer convFile.Close()
	reader := csv.NewReader(convFile)
	reader.FieldsPerRecord = -1
	reader.Comma = rune('\t')
	reader.Comment = rune('#')
	rawCSVdata, err := reader.ReadAll()
	if err != nil {
		log.Fatal(err)
	}
	conversions := make([]htmlConversion, 0)
	for _, row := range rawCSVdata {
		if len(row) == 3 {
			conversions = append(conversions, htmlConversion{row[0], row[1],
				row[2], true, ""})
		} else if len(row) == 4  {
			glossChinese, err := strconv.ParseBool(row[3])
			if err != nil {
				conversions = append(conversions, htmlConversion{row[0], row[1],
				row[2], true, ""})
			} else {
				conversions = append(conversions, htmlConversion{row[0], row[1],
					row[2], glossChinese, ""})
			}
		} else if len(row) == 5  {
			glossChinese, err := strconv.ParseBool(row[3])
			if err != nil {
				conversions = append(conversions, htmlConversion{row[0], row[1],
				row[2], true, row[4]})
			} else {
				conversions = append(conversions, htmlConversion{row[0], row[1],
					row[2], glossChinese, row[4]})
			}
		}
	}
	return conversions
}

func getCorpusConfig(c config.AppConfig) corpus.CorpusConfig {
	var excluded map[string]bool
	if len(c.CorpusDataDir()) > 0 {
		excludedFile := c.CorpusDataDir() + "/exclude.txt"
		file, err := os.Open(excludedFile)
		if err != nil {
			log.Printf("corpus.loadExcluded: Error opening excluded words file: %v, " +
				"skipping excluded words", err)
		}
		defer file.Close()
		excludedPtr, err := corpus.LoadExcluded(file)
		if err != nil {
			log.Printf("corpus.loadExcluded: Error loading excluded words file: %v, " +
				"skipping excluded words", err)
		} else {
			excluded = *excludedPtr
		}
	}
	return corpus.CorpusConfig{
		CorpusDataDir: c.CorpusDataDir(),
		CorpusDir: c.CorpusDir(),
		Excluded: excluded,
		ProjectHome: c.ProjectHome,
	}
}

func getDictionaryConfig(c config.AppConfig) dicttypes.DictionaryConfig {
	return dicttypes.DictionaryConfig{
		AvoidSubDomains: c.AvoidSubDomains(),
		DictionaryDir: c.DictionaryDir(),
	}
}

// getHTMLOutPutConfig gets the Web directory, as used for serving HTML files
func getHTMLOutPutConfig(c config.AppConfig) generator.HTMLOutPutConfig {
	domain_label := c.GetVar("Domain")
	templateHome := os.Getenv("TEMPLATE_HOME")
	//log.Printf("config.TemplateDir: templateHome: '%s'\n", templateHome)
	if len(templateHome) == 0 {
		templateHome = "html/templates"
	}
	vocabFormat := c.GetVar("VocabFormat")
	if len(vocabFormat) == 0 {
		vocabFormat = "<a title=\"%s | %s\" class=\"%s\" href=\"/words/%d.html\">%s</a>"
	}
	webDir := os.Getenv("WEB_DIR")
	if len(webDir) == 0 {
		webDir = "web"
	}
	title := c.GetVar("Title")
	if len(title) == 0 {
		title = "Chinese Notes Translation Portal"
	}
	templates := generator.NewTemplateMap(c)
	outputConfig := generator.HTMLOutPutConfig{
		Title: title,
		Templates: templates,
		ContainsByDomain: c.GetVar("ContainsByDomain"),
		Domain: domain_label,
		GoStaticDir: c.GetVar("GoStaticDir"),
		TemplateDir: templateHome,
		VocabFormat: vocabFormat,
		WebDir: webDir,
	}
	return outputConfig
}

func getIndexConfig(c config.AppConfig) index.IndexConfig {
	return index.IndexConfig{
		IndexDir: c.ProjectHome + "/index",
	}
}

// writeLibraryFiles writes HTML files for each file in the corpus.
//
// Table of contents files are also written with links including the highest
// level file pointing to the different ToC files.
func writeLibraryFiles(lib library.Library, dictTokenizer tokenizer.Tokenizer,
		outputConfig generator.HTMLOutPutConfig, corpusConfig corpus.CorpusConfig,
		indexConfig index.IndexConfig,
		wdict map[string]dicttypes.Word) error {
	libFle, err := os.Open(library.LibraryFile)
	if err != nil {
		return fmt.Errorf("writeLibraryFiles: Error opening library file: %v", err)
	}
	defer libFle.Close()
	corpora, err := lib.Loader.LoadLibrary(libFle)
	if err != nil {
		return fmt.Errorf("writeLibraryFiles, Error loading library: %v", err)
	}
	libraryOutFile := outputConfig.WebDir + "/library.html"
	analysis.WriteLibraryFile(lib, *corpora, libraryOutFile, outputConfig)
	portalDir := ""
	goStaticDir := outputConfig.GoStaticDir
	if len(goStaticDir) != 0 {
		portalDir = corpusConfig.ProjectHome + "/" + goStaticDir
		_, err := os.Stat(portalDir)
		lib.TargetStatus = "translator_portal"
		if err == nil {
			portalLibraryFile := portalDir + "/portal_library.html"
			analysis.WriteLibraryFile(lib, *corpora, portalLibraryFile, outputConfig)
		}
	}
	for _, c := range *corpora {
		outputFile := ""
		if c.Status == "public" {
			outputFile = fmt.Sprintf("%s/%s.html", outputConfig.WebDir,
					c.ShortName)
		} else if c.Status == "translator_portal" {
			outputFile = fmt.Sprintf("%s/%s.html", portalDir, c.ShortName)
		} else {
			log.Printf("library.WriteLibraryFiles: not sure what to do with status %v",
				c.Status)
			continue
		}
		r, err := os.Create(c.FileName)
		if err != nil {
			log.Fatalf("WriteHwFiles, unable to open to file %s: %v",
				c.FileName, err)
		}
		defer r.Close()
		collections, err := lib.Loader.GetCorpusLoader().LoadCorpus(r)
		if err != nil {
			return fmt.Errorf("library.WriteLibraryFiles: could not load corpus: %v", err)
		}
		_, err = analysis.WriteCorpus(*collections, outputConfig, lib.Loader,
				dictTokenizer, indexConfig, wdict)
		if err != nil {
			return fmt.Errorf("library.WriteLibraryFiles: could not open file: %v", err)
		}
		corpus := library.Corpus{c.Title, "", lib.DateUpdated, *collections}
		f, err := os.Create(outputFile)
		if err != nil {
			return fmt.Errorf("library.WriteLibraryFiles: could not open file: %v", err)
		}
		defer f.Close()
		w := bufio.NewWriter(f)
		templFile := outputConfig.TemplateDir + "/corpus-list-template.html"
		tmpl:= template.Must(template.New(
					"corpus-list-template.html").ParseFiles(templFile))
		err = tmpl.Execute(w, corpus)
		if err != nil {
			return fmt.Errorf("library.WriteLibraryFiles: could exacute template: %v", err)
		}
		w.Flush()
	}
	return nil
}

// Entry point for the chinesenotes command line tool.
// Default action is to write out all corpus entries to HTML files
func main() {
	// Command line flags
	var collectionFile = flag.String("collection", "", 
			"Enhance HTML markup and do vocabulary analysis for all the files " +
			"listed in given collection.")
	var html = flag.Bool("html", false, "Enhance HTML markup for all files " +
			"listed in data/corpus/html-conversion.csv")
	var hwFiles = flag.Bool("hwfiles", false, "Compute and write " +
			"HTML entries for each headword, writing the files to the "+
			"web/words directory.")
	var librarymeta = flag.Bool("librarymeta", false, "Top level " +
			"collection entries for the digital library.")
	var memprofile = flag.String("memprofile", "", "write memory profile to " +
			"this file")
	var sourceText = flag.String("source_text", "",
			"Analyze vocabulary for source input on the command line")
	var writeTMIndex = flag.Bool("tmindex", false, "Compute and write " +
			"translation memory index.")
	flag.Parse()

	// Minimal config for simple cases
	c := initApp()
	var wdict map[string]dicttypes.Word
	var err error
	if len(c.LUFileNames) > 0 {
		wdict, err = fileloader.LoadDictFile(c)
	} else {
		const url = "https://github.com/alexamies/chinesenotes.com/blob/master/data/words.txt?raw=true"
		wdict, err = fileloader.LoadDictURL(c, url)
	}
	if err != nil {
		log.Fatalf("Error opening dictionary: %v", err)
	}
	dictTokenizer := tokenizer.DictTokenizer{wdict}

	// Simple case, no validation done
	if len(*sourceText) > 0 {
		tokens := dictTokenizer.Tokenize(*sourceText)
		fmt.Println("Analysis of input text:")
		formatTokens(os.Stdout, tokens)
		os.Exit(0)
	} 

	outputConfig := getHTMLOutPutConfig(c)
	corpusConfig := getCorpusConfig(c)
	indexConfig := getIndexConfig(c)

	// Validate
	posFName := fmt.Sprintf("%s/%s", c.DictionaryDir(), "grammar.txt")
	posFile, err := os.Open(posFName)
	if err != nil {
		log.Fatalf("creating opening pos file %s, %v", posFName, err)
	}
	defer posFile.Close()
	posReader := bufio.NewReader(posFile)
	domainFName := fmt.Sprintf("%s/%s", c.DictionaryDir(), "topics.txt")
	domainFile, err := os.Open(domainFName)
	if err != nil {
		log.Fatalf("creating opening domain file %s, %v", domainFName, err)
	}
	domainReader := bufio.NewReader(domainFile)
	validator, err := dictionary.NewValidator(posReader, domainReader)
	if err != nil {
		log.Fatalf("creating dictionary validator: %v", err)
	}

	// Setup loader for library
	fname := c.ProjectHome + "/" + library.LibraryFile
	libraryLoader := library.NewLibraryLoader(fname, corpusConfig)

	// Validate dictionary for cases below
	err = dictionary.ValidateDict(wdict, validator)
	if err != nil {
		log.Fatalf("main: unexpected error reading headwords, %v", err)
	}

	if len(*collectionFile) > 0 {
		log.Printf("main: writing collection %s\n", *collectionFile)
		err := analysis.WriteCorpusCol(*collectionFile, libraryLoader,
				dictTokenizer, outputConfig, corpusConfig, wdict)
		if err != nil {
			log.Fatalf("error writing collection %v\n", err)
		}
	} else if *html {
		log.Printf("main: Converting all HTML files\n")
		conversions := getHTMLConversions(c)
		for _, conversion := range conversions {
			src :=  outputConfig.WebDir + "/" + conversion.SrcFile
			dest :=  outputConfig.WebDir + "/" + conversion.DestFile
			templateFile := `\N`
			if conversion.Template != `\N` {
				templateFile = outputConfig.TemplateDir + "/" + conversion.Template
			}
			log.Printf("main: input file: %s, output file: %s, template: %s\n",
				src, dest, templateFile)
			r, err := os.Create(src)
			if err != nil {
				log.Fatalf("main, could not open file: %v", err)
			}
			defer r.Close()
			text := libraryLoader.GetCorpusLoader().ReadText(r)
			tokens, results := analysis.ParseText(text, "",
					corpus.NewCorpusEntry(), dictTokenizer, getCorpusConfig(c), wdict)
			f, err := os.Create(dest)
			if err != nil {
				log.Fatalf("main, unable to write to file %s: %v", dest, err)
			}
			defer f.Close()
			err = analysis.WriteDoc(tokens, results.Vocab, f, conversion.Template,
					templateFile, conversion.GlossChinese, conversion.Title, corpusConfig, wdict)
			if err != nil {
				log.Fatalf("main, unable to write doc %s: %v", dest, err)
			}
		}
	} else if *hwFiles {
		log.Printf("main: Writing word entries for headwords\n")
		err := analysis.WriteHwFiles(libraryLoader, dictTokenizer, outputConfig,
				indexConfig, wdict)
		if err != nil {
			log.Fatalf("main, unable to write headwords: %v", err)
		}
	} else if *librarymeta {
		log.Printf("main: Writing digital library metadata\n")
		fname := c.ProjectHome + "/" + library.LibraryFile
		libraryLoader := library.NewLibraryLoader(fname, corpusConfig,)
		dateUpdated := time.Now().Format("2006-01-02")
		lib := library.Library{
			Title: "Library",
			Summary: "Top level collection in the Library",
			DateUpdated: dateUpdated,
			TargetStatus: "public",
			Loader: libraryLoader,
		}
		err := writeLibraryFiles(lib, dictTokenizer, outputConfig,
				corpusConfig, indexConfig, wdict)
		if err != nil {
			log.Fatalf("main: could not write library files: %v\n", err)
		}
	} else if *writeTMIndex {

		log.Println("main: writing translation memory index")
		err := tmindex.BuildIndexes(indexConfig.IndexDir, wdict)
		if err != nil {
			log.Fatalf("main: could not create to index file, err: %v\n", err)
		}
	} else {
		log.Println("main: writing out entire corpus")
		_, err := analysis.WriteCorpusAll(libraryLoader, dictTokenizer,
				outputConfig, indexConfig, wdict)
		if err != nil {
			log.Fatalf("main: writing out corpus, err: %v\n", err)
		}
	}

	// Memory profiling
	if *memprofile != "" {
        f, err := os.Create(*memprofile)
        if err != nil {
            log.Fatal(err)
        }
        pprof.WriteHeapProfile(f)
        f.Close()
    }
}
