export type Game = {
    name: string;
    state: GameState;
    round: RoundState;
    players: Player[];
    firstRound: Topic[];
    secondRound: Topic[];
    finalQuestion: Question;
    curQuestion: Question;
    lastAnswer: string;
    ansCorrectness: boolean;
    lastToAnswer: Player;
    guessedWrong: string[];
    paused: boolean;
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
    canVote: boolean;
    buzzBlocked: boolean;
    conn: any;
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
    RecvVote,
    PostGame,
    Error,
}

export enum RoundState {
    FirstRound,
    SecondRound,
    FinalRound,
}

export const Ping = 'ping';