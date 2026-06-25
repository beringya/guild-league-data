import { api } from "./client";

export async function latestBattleId(): Promise<number | null> {
  const battles = await api<{ items: { id: number }[] }>("/battles");
  return battles.items[0]?.id ?? null;
}
