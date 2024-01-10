export type Game = {
    state: GameState;
    round: RoundState;
    players: Player[];
    firstRound: Topic[];
    secondRound: Topic[];
    finalQuestion: Question;
    curQuestion: Question;
    lastAnswer: string;
    ansCorrectness: boolean;
    lastAnswerer: Player;
    guessedWrong: string[];
    name: string;
};

export type Player = {
    id: string;
    name: string;
    score: number;
    finalCorrect: boolean;
    finalAnswer: string;
    finalProtestors: any;
    canPick: boolean;
    canBuzz: boolean;
    canAnswer: boolean;
    canWager: boolean;
    canConfirmAns: boolean;
    buzzBlocked: boolean;
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
    RecvAnsConfirmation,
    PostGame,
    Error,
}

export enum RoundState {
	FirstRound,
	SecondRound,
	FinalRound,
}

export const Ping = 'ping';