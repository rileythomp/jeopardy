export type Game = {
    name: string;
    state: GameState;
    round: RoundState;
    players: Player[];
    firstRound: Category[];
    secondRound: Category[];
    finalQuestion: Question;
    curQuestion: Question;
    previousQuestion: string;
    previousAnswer: string;
    lastAnswer: string;
    ansCorrectness: boolean;
    lastToAnswer: Player;
    guessedWrong: string[];
    paused: boolean;
    startBuzzCountdown: boolean;
    startFinalWagerCountdown: boolean;
    startFinalAnswerCountdown: boolean;
    buzzBlocked: boolean;
};

export type Player = {
    id: string;
    name: string;
    score: number;
    finalWager: number;
    finalCorrect: boolean;
    finalAnswer: string;
    finalProtestors: any;
    canPick: boolean;
    canBuzz: boolean;
    canAnswer: boolean;
    canWager: boolean;
    canVote: boolean;
    playAgain: boolean;
    conn: any;
};

type Category = {
    title: string;
    questions: Question[];
}

export type Question = {
    category: string;
    comments: string;
    question: string;
    answer: string;
    value: number;
    canChoose: boolean;
    dailyDouble: boolean;
}

export enum GameState {
    PreGame,
    BoardIntro,
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

export type Message = {
    username: string;
    message: string;
    timestamp: number;
}