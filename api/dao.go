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

// MustInMemoryPool creates and initializes in-memory database. The database is populated
// with test data.
func MustInMemoryTestPool() (db *DbConn) {
	db = MustNewPool(":memory:")
	if _, err := db.conn.Exec(schemaSQLNew); err != nil {
		panic(err)
	}
	if err := populateTestDB(db.conn); err != nil {
		panic(err)
	}
	return
}

// Connx returns sqlx connection pool. For testing purposes.
func (db *DbConn) Connx() (conn *sqlx.DB) {
	conn = db.conn
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
	err = db.conn.Select(&tmp, "select h.orgname, cast(h.whenissued as text) as whenissued, h.xml from organizations o, history h where o.id = ? and o.name = h.orgname", id)
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

func (db *DbConn) UpdateLicenseSet(keyId string, newLicensesSet []LicenseSetItem) (err error) {
	tx, err := db.conn.Beginx()
	if err != nil {
		return
	}
	// TODO: We should check if key is valid
	// tx.Query(...)
	_, err = tx.Exec("delete from licensesets where keyid=?", keyId)
	if err != nil {
		tx.Rollback()
		return
	}
	for _, i := range newLicensesSet {
		_, err = tx.Exec("insert into licensesets (keyid, feat, ver, count, start, end, dup) values (?, ?, ?, ?, ?, ?, ?) ",
			i.KeyID, i.Feature, i.Version, i.Count, i.Start.Format("02/01/2006"), i.End.Format("02/01/2006"), i.DupGroup)
		if err != nil {
			tx.Rollback()
			return
		}
	}
	err = tx.Commit()
	return
}

func convertTimeInHistory(h historyItem) (res HistoryItem, err error) {
	res = HistoryItem{ClientName: h.ClientName, ContentXml: h.ContentXml}
	res.IssueTime, err = time.Parse("2006-01-02 15:04:05", h.IssueTime)
	return
}

const (
	schemaSQL = `CREATE TABLE organizations (id integer primary key AUTOINCREMENT, name varchar(40), contact text, comments text);
	CREATE TABLE keys (id varchar(10) primary key not null, assigned_org integer, comments text);
	CREATE TABLE history (orgname VARCHAR(40), whenissued VARCHAR(16) DEFAULT CURRENT_TIMESTAMP NOT NULL, xml TEXT);

	CREATE TABLE features (feat varchar(16) primary key, ispackage integer default 0, description text);
	CREATE TABLE pkgcontent (pkg varchar(10), feat varchar(10), primary key (pkg, feat));

	CREATE TABLE licensesets (keyid varchar(10), feat varchar(16), ver float, count integer, start DATE, end DATE, dup varchar(4), primary key (keyid, feat));

	CREATE TABLE templates (name varchar(10) not null, feat varchar(16), ver float, dup varchar(4), primary key (name, feat));
	`
	schemaSQLNew = `CREATE TABLE organizations (
		id INTEGER NOT NULL,
		name VARCHAR(40),
		contact VARCHAR,
		comments VARCHAR,
		PRIMARY KEY (id)
	);
	CREATE TABLE features (
		feat VARCHAR(16) NOT NULL,
		ispackage BOOLEAN,
		description VARCHAR,
		PRIMARY KEY (feat),
		CHECK (ispackage IN (0, 1))
	);
	CREATE TABLE keys (
		id VARCHAR(10) NOT NULL,
		assigned_org INTEGER,
		comments VARCHAR,
		PRIMARY KEY (id),
		FOREIGN KEY(assigned_org) REFERENCES organizations (id)
	);
	CREATE TABLE pkgcontent (
		pkg VARCHAR(16) NOT NULL,
		feat VARCHAR(16) NOT NULL,
		PRIMARY KEY (pkg, feat),
		FOREIGN KEY(pkg) REFERENCES features (feat),
		FOREIGN KEY(feat) REFERENCES features (feat)
	);
	CREATE TABLE templates (
		name VARCHAR(10) NOT NULL,
		feat VARCHAR(16) NOT NULL,
		ver FLOAT,
		dup VARCHAR(4),
		PRIMARY KEY (name, feat),
		FOREIGN KEY(feat) REFERENCES features (feat)
	);
	CREATE TABLE licensesets (
		keyid VARCHAR(10) NOT NULL,
		feat VARCHAR(16) NOT NULL,
		ver FLOAT,
		dup VARCHAR(4),
		count INTEGER,
		start DATE,
		"end" DATE,
		PRIMARY KEY (keyid, feat),
		FOREIGN KEY(keyid) REFERENCES keys (id),
		FOREIGN KEY(feat) REFERENCES features (feat)
	);
	CREATE TABLE history (
		id INTEGER NOT NULL,
		orgname VARCHAR(40),
		whenissued TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
		xml VARCHAR,
		PRIMARY KEY (id)
	);
	`
)

func populateTestDB(conn *sqlx.DB) (err error) {
	_, err = conn.Exec(dbTestContentSQL)
	return
}

const dbTestContentSQL = `
insert into organizations (name, contact, comments) values ('Org 1', 'Contact 1', 'No comments for "Org 1"');
insert into organizations (name, contact, comments) values ('Org 2', 'Contact 1', 'No comments for "Org 2"');

insert into features (feat, ispackage, description) values ('F1', 0, 'Descr F1');
insert into features (feat, ispackage, description) values ('F2', 0, 'Descr F2');
insert into features (feat, ispackage, description) values ('F3', 0, 'Descr F3');
insert into features (feat, ispackage, description) values ('P1', 1, 'Descr P1');

insert into keys (id, assigned_org, comments) values ('123abc', 1, 'Comments key 1');
insert into keys (id, assigned_org, comments) values ('123bbc', 1, 'Comments key 1');
insert into keys (id, assigned_org, comments) values ('123cbc', 2, 'Comments key 1');

insert into pkgcontent (pkg, feat) values ('P1', 'F1');
insert into pkgcontent (pkg, feat) values ('P1', 'F2');

insert into templates (name, feat, ver, dup) values ('Templ1', 'P1', 19.0, 'DISP');
insert into templates (name, feat, ver, dup) values ('Templ1', 'F3', 19.0, '');

insert into licensesets (keyid, feat, ver, dup, count, start, end) values ('123abc', 'P1', 19.0, 'DISP', 10, '08/07/2007', '08/07/2008');
insert into licensesets (keyid, feat, ver, dup, count, start, end) values ('123abc', 'F3', 19.0, 'DISP', 10, '08/07/2007', '08/07/2008');
insert into licensesets (keyid, feat, ver, dup, count, start, end) values ('123bbc', 'P1', 19.0, 'DISP', 20, '08/07/2007', '08/07/2008');
insert into licensesets (keyid, feat, ver, dup, count, start, end) values ('123bbc', 'F3', 19.0, 'DISP', 20, '08/07/2007', '08/07/2008');
insert into licensesets (keyid, feat, ver, dup, count, start, end) values ('123bbc', 'F1', 1.0, 'DISP', 30, '08/07/2007', '08/07/2008');

insert into history (orgname, whenissued, xml) values ('Org 1', '2018-04-26 14:24:54', '<?xml version="1.0"?><!DOCTYPE license_server><license_server port="1234" id="123ABC"></license_server>');
insert into history (orgname, whenissued, xml) values ('Org 1', '2018-04-26 14:24:54', '<?xml version="1.0"?><!DOCTYPE license_server><license_server port="1234" id="123BBC"></license_server>');
insert into history (orgname, xml) values ('Org 1',  '<?xml version="1.0"?><!DOCTYPE license_server><license_server port="1234" id="123BBC"></license_server>');
insert into history (orgname, xml) values ('Org 2',  '<?xml version="1.0"?><!DOCTYPE license_server><license_server port="1234" id="123CBC"></license_server>');
`