package main

import (
	"compress/gzip"
	"encoding/xml"
	"fmt"
	"net/http"
	"net/url"
	"os"

	"github.com/antchfx/xmlquery"
)

var (
	mirrorURL = "https://mirrors.edge.kernel.org/centos"

	repoURI = "/8-stream/BaseOS/x86_64/os"

	repoDataFolder = "repodata"
	repoMDFilename = "repomd.xml"

	repoMDURI = repoDataFolder + "/" + repoMDFilename
)

type RepoMetadata struct {
	XMLName  xml.Name `xml:"repomd"`
	Revision string   `xml:"revision"`
	Data     []Data   `xml:"data"`
}

type Data struct {
	Type     string   `xml:"type,attr"`
	Location Location `xml:"location"`
}

type Location struct {
	Href string `xml:"href,attr"`
}

type PrimaryRepoMetadata struct {
	XMLName  xml.Name  `xml:"metadata"`
	Packages []Package `xml:"package"`
}

type Package struct {
	XMLName     xml.Name        `xml:"package"`
	Name        string          `xml:"name"`
	Arch        string          `xml:"arch"`
	Version     PackageVersion  `xml:"version"`
	Summary     string          `xml:"summary"`
	Description string          `xml:"description"`
	Packager    string          `xml:"packager"`
	Url         string          `xml:"url"`
	Time        PackageTime     `xml:"time"`
	Size        PackageSize     `xml:"size"`
	Location    PackageLocation `xml:"location"`
	Format      PackageFormat   `xml:"format"`
}

type PackageVersion struct {
	XMLName xml.Name `xml:"version"`
	Epoch   string   `xml:"epoch,attr"`
	Ver     string   `xml:"ver,attr"`
	Rel     string   `xml:"rel,attr"`
}

type PackageTime struct {
	File  string `xml:"file,attr"`
	Build string `xml:"build,attr"`
}

type PackageSize struct {
	Package   string `xml:"package,attr"`
	Installed string `xml:"installed,attr"`
	Archive   string `xml:"archive,attr"`
}

type PackageLocation struct {
	XMLName xml.Name `xml:"location"`
	Href    string   `xml:"href,attr"`
}

type PackageFormat struct {
	XMLName     xml.Name           `xml:"format"`
	License     string             `xml:"license"`
	Vendor      string             `xml:"vendor"`
	Group       string             `xml:"group"`
	Buildhost   string             `xml:"buildhost"`
	HeaderRange PackageHeaderRange `xml:"header-range"`
	Requires    PackageRequires    `xml:"requires"`
	Provides    PackageProvides    `xml:"provides"`
}

type PackageHeaderRange struct {
	Start string `xml:"start,attr"`
	End   string `xml:"end,attr"`
}

type PackageProvides struct {
	XMLName xml.Name   `xml:"provides"`
	Entries []RPMEntry `xml:"entry"`
}

type PackageRequires struct {
	XMLName xml.Name   `xml:"requires"`
	Entries []RPMEntry `xml:"entry"`
}

type RPMEntry struct {
	XMLName xml.Name `xml:"entry"`
	Name    string   `xml:"name,attr"`
}

func getGzipReaderFromURL(us string) (*gzip.Reader, error) {
	u, err := url.Parse(us)
	if err != nil {
		return nil, err
	}

	resp, err := http.Get(u.String())
	if err != nil {
		return nil, err
	}

	r, err := gzip.NewReader(resp.Body)
	if err != nil {
		return nil, err
	}
	defer r.Close()

	return r, nil
}

// TODO: add support for sqlite DBs.
func GetPackagesFromRepoXMLDBURL(us string) (packages []Package, err error) {
	u, err := url.Parse(us)
	if err != nil {
		return nil, err
	}

	gr, err := getGzipReaderFromURL(u.String())
	if err != nil {
		return
	}

	doc, err := xmlquery.Parse(gr)
	if err != nil {
		return
	}

	packagesXML, err := xmlquery.QueryAll(doc, "//package")
	if err != nil {
		return
	}

	for _, v := range packagesXML {
		p := &Package{}
		err = xml.Unmarshal([]byte(v.OutputXML(true)), p)
		if err != nil {
			return
		}
		packages = append(packages, *p)

	}
	return
}

func GetDBsFromRepoMetaDataURL(us string) (DBs []Data, err error) {
	u, err := url.Parse(us)
	if err != nil {
		return nil, err
	}

	resp, err := http.Get(u.String())
	if err != nil {
		return nil, err
	}

	doc, err := xmlquery.Parse(resp.Body)
	if err != nil {
		return
	}

	DatasXML, err := xmlquery.QueryAll(doc, "//repomd/data")
	if err != nil {
		return
	}

	for _, v := range DatasXML {
		data := &Data{}
		err = xml.Unmarshal([]byte(v.OutputXML(true)), data)
		if err != nil {
			return
		}

		DBs = append(DBs, *data)
	}

	return
}

func main() {
	DBs, err := GetDBsFromRepoMetaDataURL(mirrorURL + repoURI + "/" + repoMDURI)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	for _, v := range DBs {
		DBURL := mirrorURL + repoURI + "/" + v.Location.Href
		packages, err := GetPackagesFromRepoXMLDBURL(DBURL)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		fmt.Printf("DB: %s\n", v.Type)
		fmt.Printf("DB URL: %s\n", DBURL)

		for _, v := range packages {
			fmt.Printf("\nName: %s", v.Name)
			fmt.Printf("\nVersion: %s", v.Version.Ver)
			fmt.Printf("\nSummary: %s", v.Summary)

			fmt.Printf("\nRequires:\n")
			for _, entry := range v.Format.Requires.Entries {
				fmt.Printf("%s, ", entry.Name)
			}

			fmt.Printf("\nProvides:\n")
			for _, entry := range v.Format.Provides.Entries {
				fmt.Printf("%s, ", entry.Name)
			}
			fmt.Printf("\n\n")
		}
	}
}
