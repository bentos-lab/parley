import { action, type Action } from 'easy-peasy';

export interface AudioPlayerModel {
    status: 'idle' | 'loading' | 'playing' | 'paused' | 'error';
    currentDebateId: string | null;
    src: string | null;
    currentTime: number;
    duration: number;
    playbackRate: number;
    isPlaying: boolean;
    activeChapter: number | null;
    play: Action<AudioPlayerModel, { debateId: string; src: string }>;
    pause: Action<AudioPlayerModel>;
    seek: Action<AudioPlayerModel, number>;
    setRate: Action<AudioPlayerModel, number>;
    setChapter: Action<AudioPlayerModel, number | null>;
    reset: Action<AudioPlayerModel>;
    setStatus: Action<AudioPlayerModel, AudioPlayerModel['status']>;
}

export const audioPlayerModel: AudioPlayerModel = {
    status: 'idle',
    currentDebateId: null,
    src: null,
    currentTime: 0,
    duration: 0,
    playbackRate: 1,
    isPlaying: false,
    activeChapter: null,
    play: action((state, { debateId, src }) => {
        state.status = 'loading';
        state.currentDebateId = debateId;
        state.src = src;
        state.isPlaying = true;
    }),
    pause: action((state) => {
        state.status = 'paused';
        state.isPlaying = false;
    }),
    seek: action((state, time) => {
        state.currentTime = time;
    }),
    setRate: action((state, rate) => {
        state.playbackRate = rate;
    }),
    setChapter: action((state, chapter) => {
        state.activeChapter = chapter;
    }),
    reset: action((state) => {
        state.status = 'idle';
        state.currentDebateId = null;
        state.src = null;
        state.currentTime = 0;
        state.duration = 0;
        state.playbackRate = 1;
        state.isPlaying = false;
        state.activeChapter = null;
    }),
    setStatus: action((state, status) => {
        state.status = status;
    }),
};

export interface SessionUIModel {
    sseActive: boolean;
    setSseActive: Action<SessionUIModel, boolean>;
}

export const sessionUIModel: SessionUIModel = {
    sseActive: false,
    setSseActive: action((state, active) => {
        state.sseActive = active;
    }),
};

export interface StoreModel {
    audioPlayer: AudioPlayerModel;
    sessionUI: SessionUIModel;
}

export const model: StoreModel = {
    audioPlayer: audioPlayerModel,
    sessionUI: sessionUIModel,
};
