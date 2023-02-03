package main

import (
	"fmt"
	"image"
	"image/color"

	"image/jpeg"
	"math/bits"
	"net"
	"os"
	"strconv"
	"strings"
	"time"

	"io"
	"sync"
)

const (
	HOST       = "localhost"
	PORT       = "9001"
	TYPE       = "TCP"
	BUFFERSIZE = 1024
)

func function(canal chan *image.RGBA, img image.Image, modifiedImg *image.RGBA, compteur int, q int, r int, wg *sync.WaitGroup) /* *image.RGBA*/ {
	if compteur == 0 { //Dans la première boucle on fait aussi le reste de la division euclidienne
		for i := 0; i < (compteur+1)*q+r; i++ {
			for j := 0; j < img.Bounds().Max.X; j++ {
				//Calcule de la couleur de chaque pixel
				pixel := img.At(i, j)
				originalColor := color.RGBAModel.Convert(pixel).(color.RGBA)
				r := float64(originalColor.R) * 0.92126
				g := float64(originalColor.G) * 0.5
				b := float64(originalColor.B) * 0.90722
				grey := uint8((r + g + b) / 3)
				c := color.RGBA{
					R: grey, G: grey, B: grey, A: originalColor.A,
				}
				modifiedImg.Set(i, j, c) //Modification du pixel
			}
		}
	} else { //Dans les boucles suivantes on a plus à s'occuper du reste
		for i := compteur*q + r; i < (compteur+1)*q+r; i++ {
			for j := 0; j < img.Bounds().Max.X; j++ {
				pixel := img.At(i, j)
				originalColor := color.RGBAModel.Convert(pixel).(color.RGBA)
				r := float64(originalColor.R) * 0.92126
				g := float64(originalColor.G) * 0.5
				b := float64(originalColor.B) * 0.90722
				grey := uint8((r + g + b) / 3)
				c := color.RGBA{
					R: grey, G: grey, B: grey, A: originalColor.A,
				}
				modifiedImg.Set(i, j, c)

			}
		}
	}

	wg.Done() //On indique qu'on a fini la goroutine pour le wait

	canal <- modifiedImg //On envoie dans le canal ce qui a été modifié

}

var wg sync.WaitGroup //Permet d'attendre que toutes les goroutines se terminent pour continuer le code

func main() {
	start := time.Now()

	//Connection au serveur TCP en localhost port 27007
	connection, err := net.Dial("tcp", "localhost:27007")
	if err != nil {
		panic(err)
	}
	defer connection.Close() //Attente de récupération des données avant de fermer la connexion
	fmt.Println("Connected to server, start receiving the file name and file size")
	bufferFileName := make([]byte, 64)
	bufferFileSize := make([]byte, 10)

	//Récupération des données, du nom et création du nouveau fichier
	connection.Read(bufferFileSize)
	fileSize, _ := strconv.ParseInt(strings.Trim(string(bufferFileSize), ":"), 10, 64)

	connection.Read(bufferFileName)
	fileName := strings.Trim(string(bufferFileName), ":")

	newFile, err := os.Create(fileName)

	if err != nil {
		panic(err)
	}
	defer newFile.Close() //atente de la fin d'écriture du fichier avant sa fermeture.
	var receivedBytes int64

	for { //Récupération et enregistrement des données
		if (fileSize - receivedBytes) < BUFFERSIZE {
			io.CopyN(newFile, connection, (fileSize - receivedBytes))
			connection.Read(make([]byte, (receivedBytes+BUFFERSIZE)-fileSize))
			break
		}
		io.CopyN(newFile, connection, BUFFERSIZE)
		receivedBytes += BUFFERSIZE
	}
	fmt.Println("Received file completely!")
	chemin := "./" + fileName
	//Fin de la récupération de l'image début de l'algo du filtre NB

	chemin2 := "./bonbon2.jpg"
	file, err := os.Open(chemin)

	if err != nil {
		fmt.Println("Error: File could not be opened")
		os.Exit(1)
	}

	defer file.Close()

	img, err := jpeg.Decode(file) //Ouverture de l'image
	if err != nil {
		panic(err.Error())
	}

	canal := make(chan *image.RGBA) //Création d'un canal pour communiquer les données entre les goroutines et le coeur du texte
	size := img.Bounds().Size()
	rect := image.Rect(0, 0, size.X, size.Y)
	modifiedImg := image.NewRGBA(rect)
	coeur := 12
	q, r := bits.Div(0, uint(2*img.Bounds().Max.Y), uint(coeur))

	for compteur := 0; compteur < coeur; compteur++ { //On lance 20 goroutines (2/coeurs)

		wg.Add(1)
		go function(canal, img, modifiedImg, compteur, int(q), int(r), &wg)

	}

	wg.Wait()                             //attente de la fin de toutes les goroutines
	outputFile, err := os.Create(chemin2) //creation de la nouvelle image
	if err != nil {
		fmt.Println("erreur")
	}
	defer outputFile.Close()

	err = jpeg.Encode(outputFile, <-canal, nil) //ecriture de la nouvelle image
	if err != nil {
		fmt.Println("erreur")
	}

	stop := time.Now()
	fmt.Println(stop.Sub(start))
}

type Changeable interface {
	Set(x, y int, c color.Color)
}
