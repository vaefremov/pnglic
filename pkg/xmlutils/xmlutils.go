package xmlutils

import (
	"encoding/xml"
	"fmt"
	"sort"
)

type ServerInfo struct {
	Port int
	ID   string
}

type FeatureXML struct {
	ID        string
	IsPackage bool
	Version   string
	Start     string
	End       string
	Count     int
	DupGroup  string
	Code      string
	Features  []string
}

type featureInsidePackage struct {
	ID string `xml:"id,attr"`
}

type featureSingleXML struct {
	ID       string `xml:"id,attr"`
	Version  string `xml:"version,attr"`
	Start    string `xml:"start,attr"`
	End      string `xml:"end,attr"`
	Count    int    `xml:"count,attr"`
	DupGroup string `xml:"dupgroup,attr"`
	Code     string `xml:"code,attr"`
}

type packageXML struct {
	XMLName  xml.Name               `xml:"package"`
	ID       string                 `xml:"id,attr"`
	Version  string                 `xml:"version,attr"`
	Start    string                 `xml:"start,attr"`
	End      string                 `xml:"end,attr"`
	Count    int                    `xml:"count,attr"`
	DupGroup string                 `xml:"dupgroup,attr"`
	Code     string                 `xml:"code,attr"`
	Features []featureInsidePackage `xml:"feature"`
}

type licenseServerXML struct {
	XMLName  xml.Name           `xml:"license_server"`
	Port     int                `xml:"port,attr"`
	ID       string             `xml:"id,attr"`
	Packages []packageXML       `xml:"package"`
	Features []featureSingleXML `xml:"feature"`
}

// ParseLicenseFileXML parses features contained in a valid PANGEA license file
// into the ServerInfo structure and slice of features or packages.
func ParseLicenseFileXML(content []byte) (serv ServerInfo, res []FeatureXML, err error) {
	res = []FeatureXML{}
	var server licenseServerXML
	err = xml.Unmarshal(content, &server)
	if err != nil {
		return ServerInfo{}, nil, err
	}
	for _, p := range server.Packages {
		features := []string{}
		for _, ff := range p.Features {
			features = append(features, ff.ID)
		}
		res = append(res, FeatureXML{IsPackage: true, Version: p.Version, Start: p.Start, End: p.End, Count: p.Count, DupGroup: p.DupGroup,
			Code: p.Code, ID: p.ID, Features: features})
	}
	for _, f := range server.Features {
		res = append(res, FeatureXML{IsPackage: false, Version: f.Version, Start: f.Start, End: f.End, Count: f.Count, DupGroup: f.DupGroup,
			Code: f.Code, ID: f.ID, Features: []string{}})
	}
	return ServerInfo{Port: server.Port, ID: server.ID}, res, nil
}

const featureTemplate = `    <%s
            id="%s"
            version="%s"
            start="%s"
            end="%s"
            count="%d"
            dupgroup="%s"
            code="%s" >
`

// ReorderSingleFeaturesFirst is a utility function that takes PANGEA XML license file
// as an input, reorders items so that single features come first, and
// converts it back to XML. Features and packages are sorted alphabetically.
func ReorderSingleFeaturesFirst(orig string) (res string, err error) {
	serv, features, err := ParseLicenseFileXML([]byte(orig))
	if err != nil {
		return "", err
	}
	// Sort licenses: individual features first
	sort.Slice(features, func(i, j int) bool {
		fI := features[i]
		fJ := features[j]
		if fI.IsPackage == fJ.IsPackage {
			return fI.ID < fJ.ID
		}
		return !fI.IsPackage
	})
	bodyXML := fmt.Sprintf("<?xml version=\"1.0\"?><!DOCTYPE license_server>\n\n<license_server port=\"%d\" id=\"%s\">\n", serv.Port, serv.ID)
	for _, f := range features {
		featureTag := "feature"
		if f.IsPackage {
			featureTag = "package"
		}
		bodyXML += fmt.Sprintf(featureTemplate, featureTag, f.ID, f.Version, f.Start, f.End, f.Count, f.DupGroup, f.Code)
		if f.IsPackage {
			for _, ff := range f.Features {
				bodyXML += fmt.Sprintf("        <feature id=\"%s\" />\n", ff)
			}
		}
		bodyXML += fmt.Sprintf("    </%s>\n", featureTag)
	}
	bodyXML += "</license_server>\n"
	return bodyXML, nil
}
