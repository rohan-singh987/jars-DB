package main

import (
	// "database/sql/driver"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sync"

	"github.com/jcelliott/lumber"
)

const Version = "1.0.0"

type (
	Logger interface{
		Fatal(string, ...interface{})
		Error(string, ...interface{})
		Warm(string, ...interface{})
		Info(string, ...interface{})
		Debug(string, ...interface{})
		Trace(string, ...interface{})
	}

	Driver struct{
		mutex sync.Mutex
		mutexes map[string]*sync.Mutex
		dir string
		log Logger
	}
)

type Options struct{
	Logger
}

func New(dir string, options *Options)(*Driver, error){
	dir = filepath.Clean(dir)

	opts := Options{}
	
	if options != nil{
		opts = *options
	}

	if opts.Logger == nil {
		opts.Logger = lumber.NewConsoleLogger((lumber.INFO))
	}

	driver := Driver{
		dir: dir,
		mutexes: make(map[string]*sync.Mutex) ,
		log: opts.Logger,
	}

	if _, err:= os.Stat(dir); err == nil {
		opts.Logger.Debug("Using '%s' (database already exists)\n", dir )
		return &driver, nil
	}

opts.Logger.Debug("Creating Database at '%s'...\n ", dir)
return &driver, os.MkdirAll(dir, 0755)
}

func (d *Driver) Write(collection, resource string, v interface{}) error{
	if collection == ""{
		return fmt.Errorf("Missing Collection")
	}

	if resource == ""{
		return fmt.Errorf("Missing Resource")
	}

	mutex := d.getOrCreateMutex(collection)
	mutex.Lock()
	defer mutex.Unlock()

	dir := filepath.Join(d.dir, collection)
	fnlPath := filepath.Join(dir, resource+ ".json")
	tmpPath := fnlPath + ".tmp"

	if err := os.Mkdir(dir, 0755); err != nil{
		return err
	}

	b, err := json.MarshalIndent(v, "", "\t ")
	if err!= nil{
		return err
	}

	b = append(b, byte('\n'))

	if err := ioutil.WriteFile(tmpPath, b, 0644); err != nil{
		return err
	}

	return os.Rename(tmpPath, fnlPath)
}

func (d *Driver) Read(collection, resource string, v interface{}) error{

	if collection == ""{
		return fmt.Errorf("Missing Collection")
	}

	if resource == ""{
		return fmt.Errorf("Missing Resource")
	}

	record := filepath.Join(d.dir, collection, resource)

	if _, err := stat(record); err != nil{
		return err
	}

	b, err := ioutil.ReadFile(record + ".json")
	if err != nil{
		return err
	}

	return json.Unmarshal(b, &v)
}

func  (d *Driver) ReadAll(collection string)([]string, error){

	if collection == ""{
		return nil, fmt.Errorf("Missing Collection : unable to read")
	}
	dir := filepath.Join(d.dir, collection)

	if _, err := stat(dir); err!= nil{
		return nil,err
	}
	files, _ := ioutil.ReadDir(dir)

	var records []string

	for _, file := range files{
		b, err := ioutil.ReadFile(filepath.Join(dir, file.Name()))
		if err != nil{
			return nil, err
		}

		records = append(records, string(b))
	}
	return records, nil
}

func  (d *Driver) Delete() error{

}

func (d *Driver) getOrCreateMutex(collection string) *sync.Mutex{
	 
	d.mutex.Lock()
	defer d.mutex.Unlock()
	m, ok := d.mutexes[collection]

	if !ok {
		m = &sync.Mutex{}
		d.mutexes[collection] = m
	}

	return m
} 

func stat(path string)(fi os.FileInfo, err error){
	if fi, err = os.Stat(path); os.IsNotExist(err){
		fi, err = os.Stat(path + ".json")
	}
	return
}

type Address struct{
	City string
	State string
}

type User struct{
	Name string
	Age json.Number
	Contact string
	Company string
	Address Address
}


func main(){
	dir := "./"

	db, err := New(dir, nil)
	if err != nil{
		fmt.Println("Error", err)
	}

	employees := []User{
		{"Rohan","18","9016765337", "Vit", Address{"bhopal","MadhyaPradesh"}},
		{"Ansh","20","832023232", "Vit", Address{"bhopal","MadhyaPradesh"}},
		{"Jhankar","19","2323225337", "Vit", Address{"bhopal","MadhyaPradesh"}},
		{"Shruti","25","2223444337", "Vit", Address{"bhopal","MadhyaPradesh"}},
		{"Aviral","19","987654337", "Vit", Address{"bhopal","MadhyaPradesh"}},
		{"Rohit","150","90135535337", "Vit", Address{"bhopal","MadhyaPradesh"}},
		{"Sonu","11","2242424244", "Vit", Address{"bhopal","MadhyaPradesh"}},
		{"Monu","12","982646462", "Vit", Address{"bhopal","MadhyaPradesh"}},
	}

	for _, value := range employees{
		db.Write("users", value.Name, User{
			Name: value.Name,
			Age: value.Age,
			Contact: value.Contact,
			Company: value.Company,
			Address: value.Address,

		})
	}

	records, err := db.ReadAll("users")
	if err != nil{
		fmt.Println("Error", err)
	}

	fmt.Println(records)

	allusers := []User{}

	for _, f := range records{
		employeeFound := User{}
		if err := json.Unmarshal([]byte(f), &employeeFound); err != nil{
			fmt.Println("Error", err)
		}

		allusers = append(allusers, employeeFound)
	}
	fmt.Println((allusers))

	// if err := db.Delete("user", "Rohan")

	
}