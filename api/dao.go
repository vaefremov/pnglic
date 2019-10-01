package api

import (
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
)

type HWKey struct {
	Id       string `db:"id"`
	OrgId    int    `db:"assigned_org"`
	Comments string `db:"comments"`
}

type Organization struct {
	Id       int    `db:"id"`
	Name     string `db:"name"`
	Contacts string `db:"contact"`
	Comments string `db:"comments"`
}

type historyItem struct {
	ClientName string `db:"orgname"`
	IssueTime  string `db:"whenissued"`
	ContentXml string `db:"xml"`
}

type HistoryItem struct {
	ClientName string
	IssueTime  time.Time
	ContentXml string
}

type licenseSetItem struct {
	KeyID    string  `db:"keyid"`
	Feature  string  `db:"feat"`
	Version  float32 `db:"ver"`
	Count    int     `db:"count"`
	Start    string  `db:"start"`
	End      string  `db:"end"`
	DupGroup string  `db:"dup"`
}

type LicenseSetItem struct {
	KeyID    string
	Feature  string
	Version  float32
	Count    int
	Start    time.Time
	End      time.Time
	DupGroup string
}

type Feature struct {
	Feature     string `db:"feat"`
	IsPackage   bool   `db:"ispackage"`
	Description string `db:"description"`
}

const DSN = "/Users/efremov/Projects/LIC/PNGLicenseManager/Backend/licset.sqlite"

type DbConn struct {
	conn *sqlx.DB
}

func NewPool(dsn string) (conn *DbConn, err error) {
	c, err := sqlx.Connect("sqlite3", dsn)
	return &DbConn{conn: c}, err
}

func MustNewPool(dsn string) (conn *DbConn) {
	conn, err := NewPool(dsn)
	if err != nil {
		panic(err)
	}
	return
}

func (db *DbConn) Keys() (res []HWKey, err error) {
	res = []HWKey{}
	err = db.conn.Select(&res, "select id, assigned_org, comments from keys")
	return
}

func (db *DbConn) Clients() (res []Organization, err error) {
	res = []Organization{}
	err = db.conn.Select(&res, "select id, name, contact, comments from organizations")
	return
}

func (db *DbConn) HistoryForClientId(id int) (res []HistoryItem, err error) {
	tmp := []historyItem{}
	res = []HistoryItem{}
	err = db.conn.Select(&tmp, "select h.orgname, h.whenissued, h.xml from organizations o, history h where o.id = ? and o.name = h.orgname", id)
	if err == nil {
		for _, h := range tmp {
			if newH, err := convertTimeInHistory(h); err == nil {
				res = append(res, newH)
			} else {
				return res, err
			}
		}
	}
	return
}

func (db *DbConn) LicensesSetByKeyId(keyId string) (res []LicenseSetItem, err error) {
	tmp := []licenseSetItem{}
	res = []LicenseSetItem{}
	err = db.conn.Select(&tmp, "select keyid, feat, ver, count, cast(start as varchar) as start, cast(end as varchar) as end, dup from licensesets where keyid=?", keyId)
	if err == nil {
		for _, lsi := range tmp {
			fmt.Println(lsi)
			tmpLsi := LicenseSetItem{KeyID: lsi.KeyID, Feature: lsi.Feature, Version: lsi.Version, Count: lsi.Count, DupGroup: lsi.DupGroup}
			tmpLsi.Start, err = time.Parse("02/01/2006", lsi.Start)
			if err != nil {
				return
			}
			tmpLsi.End, err = time.Parse("02/01/2006", lsi.End)
			if err != nil {
				return
			}
			res = append(res, tmpLsi)
		}
	}
	return
}

func (db *DbConn) Features() (res []Feature, err error) {
	res = []Feature{}
	err = db.conn.Select(&res, "select feat, ispackage, description from features")
	return
}

type PackageContentItem struct {
	PackageName string `db:"pkg"`
	Feature     string `db:"feat"`
}

func (db *DbConn) PackageContent(pkg string) (res []PackageContentItem, err error) {
	res = []PackageContentItem{}
	err = db.conn.Select(&res, "select pkg, feat from pkgcontent where pkg=?", pkg)
	return
}

// IsPackage checks if the given feature is a package
func (db *DbConn) IsPackage(pkg string) (res bool, err error) {
	tmp := Feature{}
	err = db.conn.Get(&tmp, "select feat, ispackage from features where feat=?", pkg)
	res = tmp.IsPackage
	return
}

func convertTimeInHistory(h historyItem) (res HistoryItem, err error) {
	res = HistoryItem{ClientName: h.ClientName, ContentXml: h.ContentXml}
	res.IssueTime, err = time.Parse("2006-01-02 15:04:05", h.IssueTime)
	return
}
