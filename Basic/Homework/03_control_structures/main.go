package main

import (
	"fmt"
	"strconv"
)

func main() {
	for {
		var chessboardSizeString string
		var chessboardSize uint64

		fmt.Println("Введите целочисленное значение размера шахматной доски большее либо равное двум:")
		_, err := fmt.Scan(&chessboardSizeString)
		if err != nil {
			fmt.Print("Введено некорректное значение размера шахматной доски.\nПопробуйте ввести значение размера еще раз.\n\n")
			continue
		} else {
			chessboardSize, err = strconv.ParseUint(chessboardSizeString, 10, 64)
			if err != nil || chessboardSize < 2 {
				fmt.Print("Введено некорректное значение размера шахматной доски.\nПопробуйте ввести значение размера еще раз.\n\n")
				continue
			}
		}

		for i := 0; i < int(chessboardSize); i++ {
			initSymbol := "#"
			nextSymbol := " "

			if i%2 != 0 {
				initSymbol = " "
				nextSymbol = "#"

			}

			for j := 0; j < int(chessboardSize/2); j++ {
				fmt.Print(initSymbol + nextSymbol)
			}

			if int(chessboardSize)%2 != 0 {
				fmt.Println(initSymbol)
			} else {
				fmt.Println("")
			}
		}
		break
	}

}
