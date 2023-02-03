package main

import (
	"fmt"
	"io"
	"net"
	"os"
	"strconv"
)

const BUFFERSIZE = 1024

func main() {
	server, err := net.Listen("tcp", "localhost:27007") //Ecoute en localhost port 27007
	if err != nil {
		fmt.Println("Error listetning: ", err)
		os.Exit(1)
	}
	defer server.Close() //On attend la fin avant la fermeture
	fmt.Println("Server started! Waiting for connections...")

	//Connection des clients
	for {
		connection, err := server.Accept()
		if err != nil {
			fmt.Println("Error: ", err)
			os.Exit(1)
		}
		fmt.Println("Client connected")
		go sendFileToClient(connection)
	}
}

func sendFileToClient(connection net.Conn) { //Transmission des fichiers au client
	fmt.Println("A client has connected!")
	defer connection.Close()          //Atennte de la fin de transmission pour fermer la connexion
	file, err := os.Open("lala.jpg") //Ouverture du fichier à transmettre
	if err != nil {
		fmt.Println(err)
		return
	}
	fileInfo, err := file.Stat()
	if err != nil {
		fmt.Println(err)
		return
	}
	//Transmission des infos du fichier (nom, taille)
	fileSize := fillString(strconv.FormatInt(fileInfo.Size(), 10), 10)
	fileName := fillString(fileInfo.Name(), 64)
	fmt.Println("Sending filename and filesize!")
	connection.Write([]byte(fileSize))
	connection.Write([]byte(fileName))
	sendBuffer := make([]byte, BUFFERSIZE)
	fmt.Println("Start sending file!")
	//Transmissiion des infos
	for {
		_, err = file.Read(sendBuffer)
		if err == io.EOF {
			break
		}
		connection.Write(sendBuffer)
	}
	fmt.Println("File has been sent, closing connection!")
	return
}

func fillString(retunString string, toLength int) string {
	for {
		lengtString := len(retunString)
		if lengtString < toLength {
			retunString = retunString + ":"
			continue
		}
		break
	}
	return retunString
}
