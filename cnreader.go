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

// Command line utility to analyze Chinese text, including corpus analysis,
// compilation of a full text search index, and mark up HTML files in reader
// style.
//
// This utility is used to generate web pages for
// https://chinesenotes.com, https://ntireader.org, and https://hbreader.org
//
// Quickstart:
//
// Supply Chinese text on the command line. Observe tokenization and matching to
// English equivalents
//
// go get github.com/alexamies/cnreader
// go run github.com/alexamies/cnreader -download_dict
// go run github.com/alexamies/cnreader -source_text="君不見黃河之水天上來"
//
// Flags:
//  -download_dict 	Download the dicitonary files from GitHub and save locally.
//  -source_text 		Analyze vocabulary for source input on the command line
//  -source_file 		Analyze vocabulary for source file and write to output.html.
//  -collection 		Enhance HTML markup and do vocabulary analysis for all the
//              		files listed in given collection.
//  -html						Enhance HTML markup for all files listed in
// 									data/corpus/html-conversion.csv
//  -hwfiles				Compute and write HTML entries for each headword, writing
// 									the files to the web/words directory.
//  -librarymeta		collection entries for the digital library.
//  -tmindex				Compute and write translation memory index.
//  -titleindex			Builds a flat index of document titles
//
// Follow instructions in the README.md file for setup.
package main

import (
	"bufio"
	"encoding/csv"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"runtime/pprof"
	"strconv"
	"strings"
	"text/template"
	"time"

	"github.com/alexamies/chinesenotes-go/bibnotes"
	"github.com/alexamies/chinesenotes-go/config"
	"github.com/alexamies/chinesenotes-go/dictionary"
	"github.com/alexamies/chinesenotes-go/dicttypes"
	"github.com/alexamies/chinesenotes-go/tokenizer"
	"github.com/alexamies/cnreader/analysis"
	"github.com/alexamies/cnreader/corpus"
	"github.com/alexamies/cnreader/generator"
	"github.com/alexamies/cnreader/index"
	"github.com/alexamies/cnreader/library"
	"github.com/alexamies/cnreader/tmindex"
)

const (
	conversionsFile = "data/corpus/html-conversion.csv"
	file2RefKey = "File2Ref"
	refNo2ParallelKey = "RefNo2ParallelKey"
	refNo2TransKey = "RefNo2Trans"
	titleIndexFN    = "documents.tsv"
)

// A type that holds the source and destination files for HTML conversion
type htmlConversion struct {
	SrcFile, DestFile, Template string
	GlossChinese                bool
	Title                       string
}

// hwWriter manages files for writing headwords to HTML
type hwWriter struct {
	outputConfig generator.HTMLOutPutConfig
	files        map[int]*os.File
}

func newHwWriter(outputConfig generator.HTMLOutPutConfig) hwWriter {
	files := make(map[int]*os.File)
	return hwWriter{
		outputConfig: outputConfig,
		files:        files,
	}
}

// OpenWriter opens the file to write HTML
func (w hwWriter) NewWriter(hwId int) io.Writer {
	filename := fmt.Sprintf("%s%s%d%s", w.outputConfig.WebDir, "/words/",
		hwId, ".html")
	f, err := os.Create(filename)
	if err != nil {
		log.Fatalf("main: Error creating file for hw.Id %d, err: %v", hwId, err)
	}
	w.files[hwId] = f
	return f
}

// CloseWriter closes the file
func (w hwWriter) CloseWriter(hwId int) {
	if f, ok := w.files[hwId]; ok {
		err := f.Close()
		if err != nil {
			log.Printf("main: CloseWriter error closing file for hw.Id %d, err: %v",
				hwId, err)
		}
	} else {
		log.Printf("main: CloseWriter could not find file for hw.Id %d", hwId)
	}
}

// Initialize the app config data
func initApp() config.AppConfig {
	return config.InitConfig()
}

// dlDictionary downloads and saves dictionary files locally.
// Also, create a config.yaml file to track it.
func dlDictionary(c config.AppConfig) error {
	const baseUrl = "https://github.com/alexamies/chinesenotes.com/blob/master/data/%s?raw=true"

	// Files to download
	luFiles := []string{"words.txt"}

	// Check if config file exists and, if not, save a fresh one
	cName := c.ProjectHome + "/config.yaml"
	_, err := os.Stat(cName)
	if os.IsNotExist(err) {
		err := saveNewConfig(cName)
		if err != nil {
			return fmt.Errorf("could not save new config file: %v", err)
		}
	} else if err != nil {
		return fmt.Errorf("could not check existence of config file: %v", err)
	} else {
		// Config is set, use it to file files to refresh
		luFiles = c.LUFileNames
	}

	// Download and save dictionary files
	files := append(luFiles, "grammar.txt")
	files = append(files, "topics.txt")
	fmt.Printf("Downloading %d files\n", len(files))
	for _, fName := range files {
		i := strings.LastIndex(fName, "/")
		var name string
		if i < 0 {
			name = fName
		} else {
			name = fName[i+1:]
		}
		url := fmt.Sprintf(baseUrl, name)
		fmt.Printf("Downloading dictionary from %s\n", url)
		resp, err := http.Get(url)
		if err != nil {
			return fmt.Errorf("GET error: %v", err)
		}
		defer resp.Body.Close()
		if resp.StatusCode > 200 {
			fmt.Printf("Could not get dictionary file %s (%d), skipping\n", name,
				resp.StatusCode)
			continue
		}
		b, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("error reading download: %v", err)
		}
		fn := c.DictionaryDir() + "/" + name
		f, err := os.Create(fn)
		if err != nil {
			return fmt.Errorf("could not create dictionary file %v", err)
		}
		defer f.Close()
		w := bufio.NewWriter(f)
		_, err = w.Write(b)
		fmt.Printf("Saved dictionary file %s\n", fName)
		if err != nil {
			return fmt.Errorf("could not write dictionary file %s : %v", fName, err)
		}
		w.Flush()
	}

	return nil
}

// saveNewConfig saves a fresh config.yaml file
func saveNewConfig(cName string) error {
	cFile, err := os.Create(cName)
	if err != nil {
		return fmt.Errorf("could not create config file %s, :%v", cName, err)
	}
	defer cFile.Close()
	const configContent = `# Generated configuration data

# Location of dictionary word files
DictionaryDir: data

# Names of dictionary files
LUFiles: words.txt

# Location for serving static resources
GoStaticDir: web

# Contains Material Design Web Go HTML templates
TemplateDir: web-resources

# Title for dynamicall enerated Go HTML pages
Title: Chinese Notes Translation Portal
`
	cWriter := bufio.NewWriter(cFile)
	_, err = cWriter.WriteString(configContent)
	if err != nil {
		return fmt.Errorf("could not write config file %s : %v", cName, err)
	}
	cWriter.Flush()
	fmt.Printf("Saved new config file %s\n", cName)
	return nil
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
			j := i + 1
			fmt.Fprintf(w, "%d. %s\n", j, ws.English)
		}
		fmt.Fprintln(w)
	}
}

// getHTMLConversions gets the list of source and destination files for HTML conversion
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
		} else if len(row) == 4 {
			glossChinese, err := strconv.ParseBool(row[3])
			if err != nil {
				conversions = append(conversions, htmlConversion{row[0], row[1],
					row[2], true, ""})
			} else {
				conversions = append(conversions, htmlConversion{row[0], row[1],
					row[2], glossChinese, ""})
			}
		} else if len(row) == 5 {
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

// getCorpusConfig loads the corpus configuration
func getCorpusConfig(c config.AppConfig) corpus.CorpusConfig {
	log.Printf("main.getCorpusConfig c.CorpusDataDir(): %s", c.CorpusDataDir())
	var excluded map[string]bool
	if len(c.CorpusDataDir()) > 0 {
		excludedFile := c.CorpusDataDir() + "/exclude.txt"
		file, err := os.Open(excludedFile)
		if err != nil {
			log.Printf("main.getCorpusConfig: Error opening excluded words file: %v, "+
				"skipping excluded words", err)
		} else {
			defer file.Close()
			excludedPtr, err := corpus.LoadExcluded(file)
			if err != nil {
				log.Printf("main.getCorpusConfig: Error loading excluded words file: %v, "+
					"skipping excluded words", err)
			} else {
				excluded = *excludedPtr
			}
		}
	}
	return corpus.NewFileCorpusConfig(c.CorpusDataDir(), c.CorpusDir(), excluded,
		c.ProjectHome)
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
		vocabFormat = `<a title="%s | %s" class="%s" href="/words/%d.html">%s</a>`
	}
	webDir := os.Getenv("WEB_DIR")
	if len(webDir) == 0 {
		webDir = "web"
	}
	title := c.GetVar("Title")
	if len(title) == 0 {
		title = "Chinese Notes Translation Portal"
	}
	match := c.GetVar("NotesReMatch")
	replace := c.GetVar("NotesReplace")
	templates := generator.NewTemplateMap(c)
	outputConfig := generator.HTMLOutPutConfig{
		Title:            title,
		Templates:        templates,
		ContainsByDomain: c.GetVar("ContainsByDomain"),
		Domain:           domain_label,
		GoStaticDir:      c.GetVar("GoStaticDir"),
		TemplateDir:      templateHome,
		VocabFormat:      vocabFormat,
		WebDir:           webDir,
		NotesReMatch:     match,
		NotesReplace:     replace,
	}
	return outputConfig
}

// getIndexConfig returns the index configuration
func getIndexConfig(c config.AppConfig) index.IndexConfig {
	return index.IndexConfig{
		IndexDir: c.ProjectHome + "/index",
	}
}

// getBibNotes initializes the BibNotesClient for bibliographic notes
func getBibNotes(cfg config.AppConfig) (bibnotes.BibNotesClient, error) {
	file2RefFN := cfg.GetVar(file2RefKey)
	if len(file2RefFN) == 0 {
		return nil, fmt.Errorf("bibnotes file2Ref not configured")
	}
	file2RefFile, err := os.Open(file2RefFN)
	if err != nil {
		return nil, fmt.Errorf("bibnotes error opening file2RefFile %s: %v",
				file2RefFN, err)
	}
	defer file2RefFile.Close()

	refNo2ParallelFN := cfg.GetVar(refNo2ParallelKey)
	if len(refNo2ParallelFN) == 0 {
		return nil, fmt.Errorf("bibnotes refNo2ParallelKey not configured")
	}
	refNo2ParallelFNFile, err := os.Open(refNo2ParallelFN)
	if err != nil {
		return nil, fmt.Errorf("bibnotes error opening refNo2TransFNFile %s: %v", 
				refNo2ParallelFN, err)
	}
	defer refNo2ParallelFNFile.Close()

	refNo2TransFN := cfg.GetVar(refNo2TransKey)
	if len(refNo2TransFN) == 0 {
		return nil, fmt.Errorf("bibnotes refNo2Trans not configured")
	}
	refNo2TransFNFile, err := os.Open(refNo2TransFN)
	if err != nil {
		return nil, fmt.Errorf("bibnotes error opening refNo2TransFNFile %s: %v", 
				refNo2TransFN, err)
	}
	defer refNo2TransFNFile.Close()

	log.Printf("Loading bib notes from %s, %s", file2RefFN, refNo2TransFN)
	client, err := bibnotes.LoadBibNotes(file2RefFile, refNo2ParallelFNFile, refNo2TransFNFile)
	if err != nil {
		return nil, fmt.Errorf("bibnotes loading error: %v", err)
	}
	return client, nil
}

// readIndex reads the index files from disk
func readIndex(indexConfig index.IndexConfig) (*index.IndexState, error) {
	fname := indexConfig.IndexDir + "/" + index.WfCorpusFile
	wfFile, err := os.Open(fname)
	if err != nil {
		return nil, fmt.Errorf("readIndex, error opening word freq file: %f", err)
	}
	defer wfFile.Close()
	wfDocFName := indexConfig.IndexDir + "/" + index.WfDocFile
	wfDocReader, err := os.Open(wfDocFName)
	if err != nil {
		return nil, fmt.Errorf("readIndex, error opening word freq doc file: %v", err)
	}
	defer wfDocReader.Close()
	indexFN := indexConfig.IndexDir + "/" + index.KeywordIndexFile
	indexWriter, err := os.Create(indexFN)
	if err != nil {
		return nil, fmt.Errorf("readIndex: Could not create file: %v", err)
	}
	defer indexWriter.Close()
	indexStore := index.IndexStore{
		WfReader:    wfFile,
		WfDocReader: wfDocReader,
		IndexWriter: indexWriter,
	}
	indexState, err := index.BuildIndex(indexConfig, indexStore)
	if err != nil {
		return nil, fmt.Errorf("readIndex: Could not build index: %v", err)
	}
	return indexState, nil
}

// writeLibraryFiles writes HTML files for each file in the corpus.
//
// Table of contents files are also written with links including the highest
// level file pointing to the different ToC files.
func writeLibraryFiles(lib library.Library, dictTokenizer tokenizer.Tokenizer,
	outputConfig generator.HTMLOutPutConfig, corpusConfig corpus.CorpusConfig,
	indexConfig index.IndexConfig,
	wdict map[string]dicttypes.Word, appConfig config.AppConfig) error {
	log.Printf("writeLibraryFiles, LibraryFile: %s", library.LibraryFile)
	libFle, err := os.Open(library.LibraryFile)
	if err != nil {
		return fmt.Errorf("writeLibraryFiles: Error opening library file: %v", err)
	}
	defer libFle.Close()
	corpora, err := lib.Loader.LoadLibrary(libFle)
	if err != nil {
		return fmt.Errorf("writeLibraryFiles, Error loading library: %v", err)
	}
	portalDir := ""
	goStaticDir := outputConfig.GoStaticDir
	if len(goStaticDir) != 0 {
		portalDir = corpusConfig.ProjectHome + "/" + goStaticDir
		_, err := os.Stat(portalDir)
		lib.TargetStatus = "public"
		if err == nil {
			libraryOutFile := portalDir + "/library.html"
			analysis.WriteLibraryFile(lib, *corpora, libraryOutFile, outputConfig)
		}
	}
	for _, c := range *corpora {
		outputFile := fmt.Sprintf("%s/%s.html", outputConfig.WebDir, c.ShortName)
		srcFileName := fmt.Sprintf("%s/%s", corpusConfig.CorpusDataDir, c.FileName)
		r, err := os.Open(srcFileName)
		if err != nil {
			log.Fatalf("writeLibraryFiles, unable to open to file %s: %v",
				c.FileName, err)
		}
		defer r.Close()
		collections, err := lib.Loader.GetCorpusLoader().LoadCorpus(r)
		if err != nil {
			return fmt.Errorf("library.WriteLibraryFiles: could not load corpus: %v", err)
		}
		log.Printf("writeLibraryFiles, loaded %d collections from corpus: %s",
			len(*collections), srcFileName)
		_, err = analysis.WriteCorpus(*collections, outputConfig, lib.Loader,
			dictTokenizer, indexConfig, wdict, appConfig, corpusConfig)
		if err != nil {
			return fmt.Errorf("library.WriteLibraryFiles: could not open file: %v", err)
		}
		corpus := library.Corpus{
			Title:       c.Title,
			Summary:     "",
			DateUpdated: lib.DateUpdated,
			Collections: *collections,
		}
		f, err := os.Create(outputFile)
		if err != nil {
			return fmt.Errorf("library.WriteLibraryFiles: could not open file: %v", err)
		}
		defer f.Close()
		w := bufio.NewWriter(f)
		templFile := outputConfig.TemplateDir + "/corpus-list-template.html"
		tmpl := template.Must(template.New(
			"corpus-list-template.html").ParseFiles(templFile))
		err = tmpl.Execute(w, corpus)
		if err != nil {
			return fmt.Errorf("library.WriteLibraryFiles: could exacute template: %v", err)
		}
		w.Flush()
	}
	return nil
}

// main is the entry point for the cnreader command line tool.
//
// The default action is to write out all corpus entries to HTML files
func main() {
	// Command line flags
	var collectionFile = flag.String("collection", "",
		"Enhance HTML markup and do vocabulary analysis for all the files "+
			"listed in given collection.")
	var downloadDict = flag.Bool("download_dict", false,
		"Download the dicitonary files from GitHub and save locally.")
	var html = flag.Bool("html", false, "Enhance HTML markup for all files "+
		"listed in data/corpus/html-conversion.csv")
	var hwFiles = flag.Bool("hwfiles", false, "Compute and write "+
		"HTML entries for each headword, writing the files to the "+
		"web/words directory.")
	var librarymeta = flag.Bool("librarymeta", false, "Top level "+
		"collection entries for the digital library.")
	var memprofile = flag.String("memprofile", "", "write memory profile to "+
		"this file.")
	var sourceFile = flag.String("source_file", "",
		"Analyze vocabulary for source file and write to output.html.")
	var sourceText = flag.String("source_text", "",
		"Analyze vocabulary for source input on the command line.")
	var writeTMIndex = flag.Bool("tmindex", false, "Compute and write "+
		"translation memory index.")
	var titleIndex = flag.Bool("titleindex", false, "Builds a flat index of "+
		"document titles.")
	flag.Parse()

	// Download latest dictionary files
	c := initApp()
	if *downloadDict {
		err := dlDictionary(c)
		if err != nil {
			log.Fatalf("Unable to download dictionary: %v", err)
		}
		os.Exit(0)
	}

	// Minimal config for simple cases
	var wdict map[string]dicttypes.Word
	var err error
	if len(c.LUFileNames) > 0 {
		wdict, err = dictionary.LoadDictFile(c)
	} else {
		const url = "https://github.com/alexamies/chinesenotes.com/blob/master/data/words.txt?raw=true"
		wdict, err = dictionary.LoadDictURL(c, url)
	}
	if err != nil {
		log.Fatalf("Error opening dictionary: %v", err)
	}
	dictTokenizer := tokenizer.DictTokenizer{
		WDict: wdict,
	}

	// Simple cases, no validation done
	if len(*sourceText) > 0 {
		tokens := dictTokenizer.Tokenize(*sourceText)
		fmt.Println("Analysis of input text:")
		formatTokens(os.Stdout, tokens)
		os.Exit(0)
	}
	outputConfig := getHTMLOutPutConfig(c)
	if len(*sourceFile) > 0 {
		r, err := os.Open(*sourceFile)
		if err != nil {
			log.Fatalf("error opening input file %s, %v", *sourceFile, err)
		}
		b, err := ioutil.ReadAll(r)
		if err != nil {
			log.Fatalf("error reading input file %s, %v", *sourceFile, err)
		}
		tokens := dictTokenizer.Tokenize(string(b))
		fName := "output.html"
		f, err := os.Create(fName)
		if err != nil {
			log.Fatalf("error creating output file %s, %v", fName, err)
		}
		defer f.Close()
		template, ok := outputConfig.Templates["texts-template.html"]
		if !ok {
			log.Fatal("no template found")
		}
		const vocabFormat = `<details><summary>%s</summary>%s %s</details>`
		err = generator.WriteDoc(tokens, f, *template, true, "Marked up page",
			vocabFormat, generator.MarkVocabSummary)
		if err != nil {
			log.Fatalf("error creating opening pos file %s, %v", fName, err)
		}
		log.Printf("Output written to %s", fName)
		os.Exit(0)
	}

	corpusConfig := getCorpusConfig(c)
	indexConfig := getIndexConfig(c)

	// Validate
	posFName := fmt.Sprintf("%s/%s", c.DictionaryDir(), "grammar.txt")
	posFile, err := os.Open(posFName)
	if err != nil {
		log.Fatalf("error opening pos file %s, %v", posFName, err)
	}
	defer posFile.Close()
	posReader := bufio.NewReader(posFile)
	domainFName := fmt.Sprintf("%s/%s", c.DictionaryDir(), "topics.txt")
	domainFile, err := os.Open(domainFName)
	if err != nil {
		log.Fatalf("error opening domain file %s, %v", domainFName, err)
	}
	domainReader := bufio.NewReader(domainFile)
	validator, err := dictionary.NewValidator(posReader, domainReader)
	if err != nil {
		log.Fatalf("error creating dictionary validator: %v", err)
	}

	// Setup loader for library
	fname := c.ProjectHome + "/" + library.LibraryFile
	libraryLoader := library.NewLibraryLoader(fname, corpusConfig)

	// Validate dictionary for cases below
	err = dictionary.ValidateDict(wdict, validator)
	if err != nil {
		log.Fatalf("main: unexpected error reading headwords, %v", err)
	}

	// Bibliographic notes client
	bibNotesClient, err := getBibNotes(c)
	if err != nil {
		log.Fatalf("main: non-fatal error, could not load bib notes: %v", err)
	}

	if len(*collectionFile) > 0 {
		log.Printf("main: writing collection %s\n", *collectionFile)
		err := analysis.WriteCorpusCol(*collectionFile, libraryLoader,
			dictTokenizer, outputConfig, corpusConfig, wdict, c)
		if err != nil {
			log.Fatalf("error writing collection %v\n", err)
		}
	} else if *html {
		conversions := getHTMLConversions(c)
		log.Printf("main: Converting %d HTML files", len(conversions))
		for _, conversion := range conversions {
			src := outputConfig.WebDir + "/" + conversion.SrcFile
			dest := outputConfig.WebDir + "/" + conversion.DestFile
			// log.Printf("main, converting file %s to %s", src, dest)
			r, err := os.Open(src)
			if err != nil {
				log.Fatalf("main, could not open file %s: %v", src, err)
			}
			defer r.Close()
			text := corpus.ReadText(r)
			tokens := dictTokenizer.Tokenize(text)
			f, err := os.Create(dest)
			if err != nil {
				log.Fatalf("main, unable to write to file %s: %v", dest, err)
			}
			defer f.Close()
			template, ok := outputConfig.Templates[conversion.Template]
			if !ok {
				log.Fatalf("template %s not found", conversion.Template)
			}
			vocabFormat := outputConfig.VocabFormat
			err = generator.WriteDoc(tokens, f, *template, conversion.GlossChinese,
				conversion.Title, vocabFormat, generator.MarkVocabLink)
			if err != nil {
				log.Fatalf("main, unable to write doc %s: %v", dest, err)
			}
		}
	} else if *hwFiles {
		log.Printf("main: Writing word entries for headwords\n")
		indexState, err := readIndex(indexConfig)
		if err != nil {
			log.Fatalf("main, unable to read index: %v", err)
		}
		vocabAnalysis, err := analysis.GetWordFrequencies(libraryLoader,
			dictTokenizer, wdict)
		if err != nil {
			log.Fatalf("main, error getting freq: %v", err)
		}
		hww := newHwWriter(outputConfig)
		hWFileDependencies := analysis.HWFileDependencies {
			Loader: libraryLoader,
			DictTokenizer: dictTokenizer,
			OutputConfig: outputConfig,
			IndexState: *indexState,
			Wdict: wdict,
			VocabAnalysis: *vocabAnalysis,
			Hww: hww,
			BibNotesClient: bibNotesClient,
		}
		err = analysis.WriteHwFiles(hWFileDependencies)
		if err != nil {
			log.Fatalf("main, unable to write headwords: %v", err)
		}
	} else if *librarymeta {
		fname := c.ProjectHome + "/" + library.LibraryFile
		log.Printf("main: Writing digital library metadata: %s\n", fname)
		libraryLoader := library.NewLibraryLoader(fname, corpusConfig)
		dateUpdated := time.Now().Format("2006-01-02")
		lib := library.Library{
			Title:        "Library",
			Summary:      "Top level collection in the Library",
			DateUpdated:  dateUpdated,
			TargetStatus: "public",
			Loader:       libraryLoader,
		}
		err := writeLibraryFiles(lib, dictTokenizer, outputConfig,
			corpusConfig, indexConfig, wdict, c)
		if err != nil {
			log.Fatalf("main: could not write library files: %v\n", err)
		}

	} else if *writeTMIndex {
		log.Println("main: writing translation memory index")
		err := tmindex.BuildIndexes(indexConfig.IndexDir, wdict)
		if err != nil {
			log.Fatalf("main: could not create tm index file, err: %v\n", err)
		}

	} else if *titleIndex {
		log.Println("main: building title index")
		fname := indexConfig.IndexDir + "/" + titleIndexFN
		f, err := os.Create(fname)
		if err != nil {
			log.Fatalf("main: could not create title index file, err: %v\n", err)
		}
		err = index.BuildDocTitleIndex(libraryLoader, f)
		if err != nil {
			log.Fatalf("main: could not build title index file, err: %v\n", err)
		}

	} else {
		log.Println("main: writing out entire corpus")
		_, err := analysis.WriteCorpusAll(libraryLoader, dictTokenizer,
			outputConfig, indexConfig, wdict, c)
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
