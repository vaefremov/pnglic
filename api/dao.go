package api

import (
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
	"github.com/pkg/errors"
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

func (db *DbConn) CreateKey(key HWKey) (err error) {
	tx, err := db.conn.Beginx()
	defer tx.Rollback()
	if err != nil {
		return errors.Wrap(err, "unable to begin transaction in CreateKey:")
	}
	tmp := Organization{}
	if err = tx.Get(&tmp, "select id, name, contact, comments from organizations where id=?", key.OrgId); err != nil {
		return fmt.Errorf("invalid org ID %d", key.OrgId)
	}
	_, err = tx.Exec("insert into keys (id, assigned_org, comments) values (?, ?, ?)", key.Id, key.OrgId, key.Comments)
	if err != nil {
		return errors.Wrap(err, "when inserting new key ID:")
	}
	err = tx.Commit()
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

// IsKeyBelongsToOrg checks if the given key ID can be found in keys table and that
// the key number is registered with the Organization identified by orgID
func (db *DbConn) IsKeyBelongsToOrg(keyID string, orgID int) (res bool, err error) {
	var numberOfKeys int
	if err = db.conn.QueryRow("select count(*) from keys where id=?", keyID).Scan(&numberOfKeys); err != nil {
		return
	}
	if numberOfKeys == 0 {
		return false, fmt.Errorf("the key id %s is invalid", keyID)
	}
	if err = db.conn.QueryRow("select count(*) from keys where id=? and assigned_org=?", keyID, orgID).Scan(&numberOfKeys); err != nil {
		return
	}
	return numberOfKeys > 0, nil
}

func (db *DbConn) LicensesSetByKeyId(keyId string) (res []LicenseSetItem, err error) {
	tmp := []licenseSetItem{}
	res = []LicenseSetItem{}
	err = db.conn.Select(&tmp, "select keyid, feat, ver, count, cast(start as varchar) as start, cast(end as varchar) as end, dup from licensesets where keyid=?", keyId)
	if err == nil {
		for _, lsi := range tmp {
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

// Features returns the full list of features
func (db *DbConn) Features() (res []Feature, err error) {
	res = []Feature{}
	err = db.conn.Select(&res, "select feat, ispackage, description from features")
	return
}

// CreateOrUpdateFeature check is the feature with the given name exists, update feature in this case,
// creates a new one otherwise
func (db *DbConn) CreateOrUpdateFeature(name string, description string, isPackage bool) (upd bool, err error) {
	tx, err := db.conn.Beginx()
	if err != nil {
		tx.Rollback()
		return
	}
	f := Feature{}
	err = tx.Get(&f, "select feat, ispackage, description from features where feat=?", name)
	if err != nil {
		_, err = tx.Exec("insert into features (feat, ispackage, description) values (?, ?, ?)", name, isPackage, description)
		upd = false
	} else {
		_, err = tx.Exec("update features set ispackage=?, description=? where feat=?", isPackage, description, name)
		upd = true
	}
	if err != nil {
		tx.Rollback()
		return
	}
	err = tx.Commit()
	return
}

// DeleteFeature deletes feature with the specified name
func (db *DbConn) DeleteFeature(name string) (err error) {
	tx, err := db.conn.Beginx()
	if err != nil {
		tx.Rollback()
		return
	}
	_, err = tx.Exec("delete from features where feat=?", name)
	if err != nil {
		tx.Rollback()
		return
	}
	err = tx.Commit()
	return
}

// SetPackageContent removes everything from the specified package and inserts all the features
// from the featureNames list
func (db *DbConn) SetPackageContent(featureNames []string, packageName string) (err error) {
	tx, err := db.conn.Beginx()
	if err != nil {
		tx.Rollback()
		return
	}
	_, err = tx.Exec("delete from pkgcontent where pkg=?", packageName)
	if err != nil {
		tx.Rollback()
		return
	}
	for _, f := range featureNames {
		_, err = tx.Exec("insert into pkgcontent (pkg, feat) values (?, ?)", packageName, f)
		if err != nil {
			tx.Rollback()
			return
		}
	}
	err = tx.Commit()
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

// AddToHistory adds license file to the history track.Client name is deduced from the
// ID
func (db *DbConn) AddToHistory(orgID int, when time.Time, fileContent string) (err error) {
	tx, err := db.conn.Beginx()
	if err != nil {
		return
	}
	// Get the client name from Id
	var orgName struct {
		Name string `db:"name"`
	}
	err = tx.Get(&orgName, "select name from organizations where id=?", orgID)
	if err != nil {
		tx.Rollback()
		return
	}
	_, err = tx.Exec("insert into history (orgname, whenissued, xml) values (?, ?, ?)", orgName.Name, when.Format("2006-01-02 15:04:05"), fileContent)
	if err != nil {
		tx.Rollback()
		return
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
insert into features (feat, ispackage, description) values ('F4', 0, 'Descr F4');
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
