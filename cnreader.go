/* 
Command line utility to mark up HTML files with Chinese notes.
 */
package main

import (
	"bufio"
	"encoding/csv"
	"flag"
	"fmt"
	"github.com/alexamies/chinesenotes-go/config"
	"github.com/alexamies/chinesenotes-go/dictionary"
	"github.com/alexamies/chinesenotes-go/fileloader"
	"github.com/alexamies/chinesenotes-go/tokenizer"
	"github.com/alexamies/cnreader/analysis"
	"github.com/alexamies/cnreader/corpus"
	"github.com/alexamies/cnreader/index"
	"github.com/alexamies/cnreader/library"
	"github.com/alexamies/cnreader/tmindex"
	"log"
	"os"
	"runtime/pprof"
	"strconv"
	"time"
)

const conversionsFile = "data/corpus/html-conversion.csv"

var (
	// Loaded from disk in contrast to partially ready and still accumulating data
	completeDF index.DocumentFrequency
	excluded map[string]bool
	projectHome string
)

// A type that holds the source and destination files for HTML conversion
type htmlConversion struct {
	SrcFile, DestFile, Template string
	GlossChinese bool
	Title string
}

func init() {
	projectHome = os.Getenv("CNREADER_HOME")
	log.Printf("config.init: projectHome: '%s'\n", projectHome)
	if len(projectHome) == 0 {
		projectHome = "."
	}
	excluded = corpus.LoadExcluded(getCorpusConfig())
	df, err := index.ReadDocumentFrequency(getIndexConfig())
	if err != nil {
		log.Println("index.init Error reading document frequency continuing")
	}
	completeDF = df
}

// Gets the list of source and destination files for HTML conversion
func getHTMLConversions() []htmlConversion {
	log.Printf("GetHTMLConversions: projectHome: '%s'\n", projectHome)
	conversionsFile := projectHome + "/" + conversionsFile
	convFile, err := os.Open(conversionsFile)
	if err != nil {
		log.Fatal("config.GetHTMLConversions fatal error: ", err)
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

func getCorpusConfig() corpus.CorpusConfig {
	return corpus.CorpusConfig{
		CorpusDataDir: config.CorpusDataDir(),
		CorpusDir: config.CorpusDir(),
		Excluded: excluded,
		ProjectHome: projectHome,
	}
}

func getDictionaryConfig() analysis.DictionaryConfig {
	return analysis.DictionaryConfig{
		AvoidSubDomains: config.AvoidSubDomains(),
		DictionaryDir: config.DictionaryDir(),
	}
}

// Gets the Web directory, as used for serving HTML files
func getHTMLOutPutConfig() corpus.HTMLOutPutConfig {
	domain_label := config.GetVar("Domain")
	templateHome := os.Getenv("TEMPLATE_HOME")
	//log.Printf("config.TemplateDir: templateHome: '%s'\n", templateHome)
	if len(templateHome) == 0 {
		templateHome = "html/templates"
	}
	vocabFormat := config.GetVar("VocabFormat")
	if len(vocabFormat) == 0 {
		vocabFormat = "<a title=\"%s | %s\" class=\"%s\" href=\"/words/%d.html\">%s</a>"
	}
	webDir := os.Getenv("WEB_DIR")
	if len(webDir) == 0 {
		webDir = "web"
	}
	outputConfig := corpus.HTMLOutPutConfig{
		ContainsByDomain: config.GetVar("ContainsByDomain"),
		Domain: domain_label,
		GoStaticDir: config.GetVar("GoStaticDir"),
		TemplateDir: templateHome,
		VocabFormat: vocabFormat,
		WebDir: webDir,
	}
	return outputConfig
}

func getIndexConfig() index.IndexConfig {
	return index.IndexConfig{
		IndexDir: projectHome + "/index",
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

	outputConfig := getHTMLOutPutConfig()
	corpusConfig := getCorpusConfig()
	indexConfig := getIndexConfig()

	// Read headwords and validate
	posFName := fmt.Sprintf("%s/%s", config.DictionaryDir(), "grammar.txt")
	posFile, err := os.Open(posFName)
	if err != nil {
		log.Fatalf("creating opening pos file %s, %v", posFName, err)
	}
	defer posFile.Close()
	posReader := bufio.NewReader(posFile)
	domainFName := fmt.Sprintf("%s/%s", config.DictionaryDir(), "topics.txt")
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
	fname := projectHome + "/" + library.LibraryFile
	fileLibraryLoader := library.FileLibraryLoader{
		FileName: fname,
		Config: corpusConfig,
	}

	wdict, err := fileloader.LoadDictFile(config.LUFileNames())
	if err != nil {
		log.Fatalf("Error opening dictionary: %v", err)
	}
	err = dictionary.ValidateDict(wdict, validator,)
	if err != nil {
		log.Fatalf("main: unexpected error reading headwords, %v", err)
	}
	dictTokenizer := tokenizer.DictTokenizer{wdict}

	if (*collectionFile != "") {
		log.Printf("main: Analyzing collection %s\n", *collectionFile)
		analysis.WriteCorpusCol(*collectionFile, fileLibraryLoader,
				dictTokenizer, outputConfig, corpusConfig, wdict)
	} else if *html {
		log.Printf("main: Converting all HTML files\n")
		conversions := getHTMLConversions()
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
					corpus.NewCorpusEntry(), dictTokenizer, getCorpusConfig(), wdict)
			analysis.WriteDoc(tokens, results.Vocab, dest, conversion.Template,
					templateFile, conversion.GlossChinese, conversion.Title, corpusConfig, wdict)
		}
	} else if *hwFiles {
		log.Printf("main: Writing word entries for headwords\n")
		analysis.WriteHwFiles(fileLibraryLoader, dictTokenizer, outputConfig,
				corpusConfig, indexConfig, wdict)
	} else if *librarymeta {
		log.Printf("main: Writing digital library metadata\n")
		fname := projectHome + "/" + library.LibraryFile
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
		analysis.WriteLibraryFiles(lib, dictTokenizer, outputConfig, corpusConfig,
				indexConfig, wdict)
	} else if *writeTMIndex {

		log.Println("main: Writing translation memory index")
		err := tmindex.BuildIndexes(indexConfig.IndexDir, wdict)
		if err != nil {
			log.Fatalf("Could not create to index file, err: %v\n", err)
		}
	} else {
		log.Println("main: Writing out entire corpus")
		analysis.WriteCorpusAll(fileLibraryLoader, dictTokenizer, outputConfig,
				corpusConfig, indexConfig, wdict)
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
