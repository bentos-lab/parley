import { z } from 'zod';

const RawAgentSchema = z.object({
    id: z.string(),
    name: z.string(),
    stance: z.string(),
    voice_name: z.string(),
});

export const AgentSchema = RawAgentSchema.transform((a) => ({
    id: a.id,
    name: a.name,
    stance: a.stance,
    voiceName: a.voice_name,
}));

export const RoundSchema = z.object({
    agent_id: z.string(),
    message: z.string(),
});

export const RawDebateSchema = z.object({
    id: z.string(),
    name: z.string(),
    normalized_name: z.string(),
    topic: z.string(),
    agents: z.array(AgentSchema),
    rounds: z.array(RoundSchema),
    tts_provider: z.string(),
});

export const DebateSchema = RawDebateSchema.transform((d) => ({
    id: d.id,
    name: d.name,
    normalizedName: d.normalized_name,
    topic: d.topic,
    agents: d.agents,
    rounds: d.rounds,
    ttsProvider: d.tts_provider,
}));

export type RawDebate = z.infer<typeof RawDebateSchema>;
export type RawDebateResponse = Omit<RawDebate, 'id'> & Partial<Pick<RawDebate, 'id'>>;

export const DebateSummarySchema = z.object({
    id: z.string(),
    name: z.string(),
    topic: z.string(),
});

export const DebateSummaryListSchema = z.array(DebateSummarySchema);

export const CreateDebateResponseSchema = z.object({
    name: z.string(),
    id: z.string(),
});

export const RoundResponseSchema = z.object({
    agent_id: z.string(),
    content: z.string(),
});
