echo '\ncurl -i -X GET http://localhost:1576/align/\n\n'
curl -i -X GET http://localhost:1576/align
echo '\ncurl -i -X GET http://localhost:1576/align?text=Biotechnology/\n\n'
curl -i -X GET http://localhost:1576/align?text=Biotechnology
echo '\ncurl -i -X GET http://localhost:1576/align?area=Science&year=6,7,8&text=Biotechnology\n\n'
curl -i -X GET "http://localhost:1576/align?area=Science&year=6,7,8&text=Biotechnology"
echo '\ncurl -i -X GET http://localhost:1576/align?area=Science&year=6,7,8&text=This%20is%20biotechnology%20%26%20geology%3B%20and%20that%27s%20that.%20%22Or%20so%20I%20have%20read.%22%0AHello%2C%20curriculum.'
curl -i -X GET "http://localhost:1576/align?area=Science&year=6,7,8&text=This%20is%20biotechnology%20%26%20geology%3B%20and%20that%27s%20that.%20%22Or%20so%20I%20have%20read.%22%0AHello%2C%20curriculum."
echo '\ncurl -i -X GET http://localhost:1576/index?search=understanding\n\n'
curl -i -X GET http://localhost:1576/index?search=understanding
