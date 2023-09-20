package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"math/rand"
	"os"
	"strconv"
	"strings"
)

type Card struct {
	Term     string `json:"term"`
	Def      string `json:"definition"`
	Mistakes int    `json:"mistakes"`
}

type CardDeck map[string]Card

type ImportFormat struct {
	Flashcards []Card `json:"flashcards"`
}

type ConsoleLogger struct {
	messages []string
	scanner  *bufio.Scanner
}

func (cl *ConsoleLogger) Println(message string) {
	cl.messages = append(cl.messages, message)
	fmt.Println(message)
}

func (cl *ConsoleLogger) Scan() string {
	message, _ := scanString(cl.scanner)
	cl.messages = append(cl.messages, "> "+message)
	return message
}

func scanString(s *bufio.Scanner) (string, error) {
	if s.Scan() {
		return strings.TrimSpace(s.Text()), nil
	}
	err := s.Err()
	if err == nil {
		err = io.EOF
	}
	return "", err
}

func readCard(card *Card, cl *ConsoleLogger, cards CardDeck) {
	cl.Println("The card:")
getterm:
	card.Term = cl.Scan()
	errTerm := cardExists(cards, *card)
	if errTerm != "" {
		cl.Println(errTerm)
		goto getterm
	}
	cl.Println("The definition of the card:")
getdef:
	card.Def = cl.Scan()
	errDef := cardExists(cards, *card)
	if errDef != "" {
		cl.Println(errDef)
		goto getdef
	}
}

func cardExists(cards CardDeck, card Card) string {
	_, ok := cards[card.Term]
	if ok {
		return fmt.Sprintf("The term \"%s\" already exists. Try again:", card.Term)
	}
	for _, val := range cards {
		if val.Def == card.Def {
			return fmt.Sprintf("The definition \"\" already exists. Trye again: ", val.Def)
		}
	}
	return ""
}

func (deck CardDeck) addCard(cl *ConsoleLogger) {
	var card Card
	readCard(&card, cl, deck)
	deck[card.Term] = card
	cl.Println(fmt.Sprintf("The pair (\"%s\":\"%s\") has been added.", card.Term, card.Def))
}

func (deck CardDeck) removeCard(cl *ConsoleLogger) {
	cl.Println("Which card?")
	answer := cl.Scan()
	_, ok := deck[answer]
	if !ok {
		cl.Println(fmt.Sprintf("Can't remove \"%s\": there is no such card.", answer))
		return
	}
	delete(deck, answer)
	cl.Println("The card has been removed.")
}

func (deck CardDeck) exportCards(cl *ConsoleLogger) {
	cl.Println("File name:")
	fileName := cl.Scan()
	file, _ := os.OpenFile(fileName, os.O_RDWR|os.O_CREATE, 766)
	numCards := len(deck)
	cards := make([]Card, 0, numCards)
	for _, val := range deck {
		cards = append(cards, val)
	}
	exportData := ImportFormat{
		Flashcards: cards,
	}
	jsonData, _ := json.MarshalIndent(exportData, "", " ")
	_, _ = file.Write(jsonData)
	cl.Println(fmt.Sprintf("%d cards have been saved.", numCards))
}

func (deck CardDeck) exportCardsFlag(cl *ConsoleLogger, exportFile *string) {
	file, _ := os.OpenFile(*exportFile, os.O_RDWR|os.O_CREATE, 766)
	numCards := len(deck)
	cards := make([]Card, 0, numCards)
	for _, val := range deck {
		cards = append(cards, val)
	}
	exportData := ImportFormat{
		Flashcards: cards,
	}
	jsonData, _ := json.MarshalIndent(exportData, "", " ")
	_, _ = file.Write(jsonData)
	cl.Println(fmt.Sprintf("%d cards have been saved.", numCards))
}

func (deck CardDeck) importCards(cl *ConsoleLogger) {
	cl.Println("File name:")
	fileName := cl.Scan()
	byteValue, err := os.ReadFile(fileName)
	if err != nil {
		cl.Println("File not found.")
		return
	}
	var importData ImportFormat
	err = json.Unmarshal(byteValue, &importData)
	if err != nil {
		return
	}
	for _, card := range importData.Flashcards {
		deck[card.Term] = card
	}
	cl.Println(fmt.Sprintf("%d cards have been loaded", len(importData.Flashcards)))
}

func (deck CardDeck) importCardsFlag(cl *ConsoleLogger, importFile *string) {
	byteValue, err := os.ReadFile(*importFile)
	if err != nil {
		cl.Println("File not found.")
		return
	}
	var importData ImportFormat
	err = json.Unmarshal(byteValue, &importData)
	if err != nil {
		return
	}
	for _, card := range importData.Flashcards {
		deck[card.Term] = card
	}
	cl.Println(fmt.Sprintf("%d cards have been loaded.", len(importData.Flashcards)))
}

func (deck CardDeck) getCardWithDef(answer string) (Card, error) {
	for _, card := range deck {
		if card.Def == answer {
			return card, nil
		}
	}
	return Card{}, errors.New("Card with this Definition doesn't exists.")
}

func (deck CardDeck) checkDefinition(card Card, cl *ConsoleLogger) {
	cl.Println(fmt.Sprintf("Print the definition of \"%s\"", card.Term))
	answer := cl.Scan()
	if answer == card.Def {
		cl.Println("Correct!")
		return
	}
	card.Mistakes += 1
	deck[card.Term] = card
	anotherCard, err := deck.getCardWithDef(answer)
	if err != nil {
		cl.Println(fmt.Sprintf("Wrong. The right answer is \"%s\".", card.Def))
		return
	}
	cl.Println(fmt.Sprintf("Wrong. The right answer is \"%s\", but your definition is correct for \"%s\".\n", card.Def, anotherCard.Term))
}

func (deck CardDeck) ask(cl *ConsoleLogger) {
	var count int
	cl.Println("How many times to ask?")
	t := cl.Scan()
	count, _ = strconv.Atoi(t)
	for i := 0; i < count; i++ {
		deck.checkDefinition(deck.randomCard(), cl)
	}
}

func (deck CardDeck) randomCard() Card {
	k := rand.Intn(len(deck))
	i := 0
	for _, x := range deck {
		if i == k {
			return x
		}
		i++
	}
	panic("unreachable")
}

func (deck CardDeck) resetStats(cl *ConsoleLogger) {
	for _, card := range deck {
		card.Mistakes = 0
		deck[card.Term] = card
	}
	cl.Println("Card statistics have been reset.")
}

func (deck CardDeck) hardestCard(cl *ConsoleLogger) {
	var stats []int
	for _, card := range deck {
		stats = append(stats, card.Mistakes)
	}
	maxError := getMax(stats)
	if maxError == 0 {
		cl.Println("There are no cards with errors.")
		return
	}
	var errored []string
	for _, card := range deck {
		if card.Mistakes == maxError {
			errored = append(errored, card.Term)
		}
	}
	if len(errored) == 1 {
		cl.Println(fmt.Sprintf("The hardest card is \"%s\". You have %d errors answering it.", errored[0], maxError))
		return
	}
	result := strings.Join(errored, "\", \"")
	cl.Println(fmt.Sprintf("The hardest cards are \"%s\". You have %d errors answering them", result, maxError))

}

func getMax(array []int) int {
	if len(array) == 0 {
		return 0
	}
	m := array[0]
	for _, el := range array {
		if m < el {
			m = el
		}
	}
	return m
}

func getAction(cl *ConsoleLogger) string {
	cl.Println("Input the action (add, remove, import, export, ask, exit, log, hardest card, reset stats):")
	action := cl.Scan()
	return action
}

func (cl *ConsoleLogger) createLogFile() {
	cl.Println("File name:")
	fileName := cl.Scan()
	if fileName == "" {
		return
	}
	file, _ := os.OpenFile(fileName, os.O_RDWR|os.O_CREATE, 766)
	sb := strings.Builder{}
	for _, message := range cl.messages {
		sb.WriteString(fmt.Sprintf("%s\n", message))
	}
	_, _ = file.WriteString(sb.String())
	cl.Println("The log has been saved.")
}

func main() {
	importFile := flag.String("import_from", "", "Import from file")
	exportFile := flag.String("export_to", "", "Export to file")

	flag.Parse()
	scanner := bufio.NewScanner(os.Stdin)
	cl := &ConsoleLogger{
		scanner: scanner,
	}
	cards := CardDeck{}

	if *importFile != "" {
		cards.importCardsFlag(cl, importFile)
	}

	isRunning := true

	for isRunning {
		switch getAction(cl) {
		case "add":
			cards.addCard(cl)
		case "remove":
			cards.removeCard(cl)
		case "export":
			cards.exportCards(cl)
		case "import":
			cards.importCards(cl)
		case "ask":
			cards.ask(cl)
		case "log":
			cl.createLogFile()
		case "hardest card":
			cards.hardestCard(cl)
		case "reset stats":
			cards.resetStats(cl)
		case "exit":
			if *exportFile != "" {
				cards.exportCardsFlag(cl, exportFile)
			}
			fmt.Println("Bye bye!")
			isRunning = false
			break
		default:
			fmt.Println("Unknown action")
		}
	}
}
