package internal

type DemoQuery struct {
}

type DemoResult struct {
}

type Demo cqrs.CommandHandler[*DemoQuery, *DemoResult]

type demo struct {
}
