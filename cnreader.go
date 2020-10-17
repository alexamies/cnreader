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
// Command line utility to analyze corpus and mark up HTML files.
//
package main

import (
	"bufio"
	"encoding/csv"
	"flag"
	"fmt"
	"log"
	"os"
	"runtime/pprof"
	"strconv"
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

// getDocFreq gets the word document frequency
//
// If it cannot be read from file, it will be computed from the corpus
func getDocFreq(c config.AppConfig) (*index.DocumentFrequency, error) {
	fname := c.ProjectHome + "/" + library.LibraryFile
	corpusConfig := getCorpusConfig(c)
	fileLibraryLoader := library.FileLibraryLoader{
		FileName: fname,
		Config: corpusConfig,
	}
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
		df, err := analysis.GetDocFrequencies(fileLibraryLoader, dictTokenizer,
				corpusConfig, wdict)
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
		df, err = analysis.GetDocFrequencies(fileLibraryLoader, dictTokenizer,
				corpusConfig, wdict)
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
	excluded := corpus.LoadExcluded(getCorpusConfig(c))
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
	var writeTMIndex = flag.Bool("tmindex", false, "Compute and write " +
		"translation memory index.")
	var librarymeta = flag.Bool("librarymeta", false, "Top level " +
		"collection entries for the digital library.")
	var memprofile = flag.String("memprofile", "", "write memory profile to " +
				"this file")
	flag.Parse()

	c := initApp()

	outputConfig := getHTMLOutPutConfig(c)
	corpusConfig := getCorpusConfig(c)
	indexConfig := getIndexConfig(c)

	// Read headwords and validate
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
		log.Fatalf("creatting dictionary validator: %v", err)
	}

	// Setup loader for library
	fname := c.ProjectHome + "/" + library.LibraryFile
	fileLibraryLoader := library.FileLibraryLoader{
		FileName: fname,
		Config: corpusConfig,
	}

	wdict, err := fileloader.LoadDictFile(c)
	if err != nil {
		log.Fatalf("Error opening dictionary: %v", err)
	}
	err = dictionary.ValidateDict(wdict, validator)
	if err != nil {
		log.Fatalf("main: unexpected error reading headwords, %v", err)
	}
	dictTokenizer := tokenizer.DictTokenizer{wdict}

	if (*collectionFile != "") {
		log.Printf("main: writing collection %s\n", *collectionFile)
		err := analysis.WriteCorpusCol(*collectionFile, fileLibraryLoader,
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
			text := fileLibraryLoader.GetCorpusLoader().ReadText(src)
			tokens, results := analysis.ParseText(text, "",
					corpus.NewCorpusEntry(), dictTokenizer, getCorpusConfig(c), wdict)
			analysis.WriteDoc(tokens, results.Vocab, dest, conversion.Template,
					templateFile, conversion.GlossChinese, conversion.Title, corpusConfig, wdict)
		}
	} else if *hwFiles {
		log.Printf("main: Writing word entries for headwords\n")
		analysis.WriteHwFiles(fileLibraryLoader, dictTokenizer, outputConfig,
				corpusConfig, indexConfig, wdict)
	} else if *librarymeta {
		log.Printf("main: Writing digital library metadata\n")
		fname := c.ProjectHome + "/" + library.LibraryFile
		fileLibraryLoader := library.FileLibraryLoader{
			FileName: fname,
			Config: corpusConfig,
		}
		dateUpdated := time.Now().Format("2006-01-02")
		lib := library.Library{
			Title: "Library",
			Summary: "Top level collection in the Library",
			DateUpdated: dateUpdated,
			TargetStatus: "public",
			Loader: fileLibraryLoader,
		}
		err := analysis.WriteLibraryFiles(lib, dictTokenizer, outputConfig,
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
		err := analysis.WriteCorpusAll(fileLibraryLoader, dictTokenizer, outputConfig,
				corpusConfig, indexConfig, wdict)
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
