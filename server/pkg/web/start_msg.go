package web

var asciibanner = `

 ____              _    _                  
|  _ \            | |  | |                 
| |_) | ___   ___ | | _| |_ _   _ _ __ ___ 
|  _ < / _ \ / _ \| |/ / __| | | | '__/ _ \
| |_) | (_) | (_) |   <| |_| |_| | | |  __/
|____/ \___/ \___/|_|\_\\__|\__,_|_|  \___|

`

func GetStartMessage(port string) string {
	return asciibanner + "\n" +
		"Running on port: " + port + "\n"
}