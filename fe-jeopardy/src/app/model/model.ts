export type Game = {
    name: string;
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
    startBuzzCountdown: boolean;
    startFinalWagerCountdown: boolean;
    startFinalAnswerCountdown: boolean;
    buzzBlocked: boolean;
    disconnected: boolean;
    officialAnswer: string;
    penalty: boolean;
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
    canDispute: boolean;
    playAgain: boolean;
    conn: any;
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
    RecvVote,
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

export let ServerUnavailableMsg = 'Sorry, Jeopardy is not available right now. Please try again later.'