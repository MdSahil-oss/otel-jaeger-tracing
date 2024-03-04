import data from './data.json' assert { type: 'json' };

db = connect("localhost:27017", "root", "password");

db.data.insertMany(data)
