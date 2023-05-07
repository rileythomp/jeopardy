export type Game = {
    state: GameState;
    players: Player[];
    firstRound: Topic[];
    secondRound: Topic[];
    finalRound: Question;
    curQuestion: Question;
};

export type Player = {
    id: string;
    name: string;
    score: number;
    canPick: boolean;
    canBuzz: boolean;
    canAnswer: boolean;
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
    RecvAns,
    RecvFinal,
    PostGame,
    Error,
}