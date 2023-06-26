export type Game = {
    state: GameState;
    round: RoundState;
    players: Player[];
    firstRound: Topic[];
    secondRound: Topic[];
    finalQuestion: Question;
    curQuestion: Question;
};

export type Player = {
    id: string;
    name: string;
    score: number;
    canPick: boolean;
    canBuzz: boolean;
    canAnswer: boolean;
    canWager: boolean;
};

type Topic = {
    title: string;
    questions: Question[];
}

export type Question = {
    question: string;
    answer: string;
    value: number;
    canChoose: boolean;
    dailyDouble: boolean;
}

export enum GameState {
    PreGame,
    RecvPick,
    RecvBuzz,
    RecvWager,
    RecvAns,
    PostGame,
    Error,
}

export enum RoundState {
	FirstRound,
	SecondRound,
	FinalRound,
}