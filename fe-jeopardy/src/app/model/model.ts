export type User = {
    imgUrl: string;
    authenticated: boolean;
    name: string;
    email: string;
    dateJoined: string;
    public: boolean;
}

export type Game = {
    name: string;
    code: string;
    state: GameState;
    round: RoundState;
    players: Player[];
    firstRound: Category[];
    secondRound: Category[];
    finalQuestion: Question;
    curQuestion: Question;
    ansCorrectness: boolean;
    guessedWrong: string[];
    paused: boolean;
    startFinalWagerCountdown: boolean;
    startFinalAnswerCountdown: boolean;
    buzzBlocked: boolean;
    disconnected: boolean;
    officialAnswer: string;
    penalty: boolean;
    pickTimeout: number;
    buzzTimeout: number;
    answerTimeout: number;
    wagerTimeout: number;
    finalAnswerTimeout: number;
    finalWagerTimeout: number;
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
    canDispute: boolean;
    playAgain: boolean;
    conn: any;
    imgUrl: string;
};

type Category = {
    title: string;
    questions: Question[];
}

export type Answer = {
    player: Player;
    answer: string;
    correct: boolean;
    hasDisputed: boolean;
    overturned: boolean;
}

export type Question = {
    category: string;
    comments: string;
    question: string;
    answer: string;
    value: number;
    canChoose: boolean;

    curAns: Answer;
    answers: Answer[];
    curDisputed: Answer;
}

export enum GameState {
    PreGame,
    BoardIntro,
    RecvPick,
    RecvBuzz,
    RecvWager,
    RecvAns,
    RecvDispute,
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

export type Reaction = {
    username: string;
    reaction: string;
    timestamp: number;
    randPos: number;
}

export let ServerUnavailableMsg = 'Sorry, Jeopardy is not available right now. Please try again later.'