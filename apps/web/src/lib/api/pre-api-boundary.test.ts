import { readdirSync, readFileSync, statSync } from 'node:fs';
import { join } from 'node:path';
import { describe, expect, it } from 'vitest';

const guardedDirectories = ['src/app', 'src/components', 'src/fixtures'];
const blockedRuntimeMarkers = [
  'fetch(',
  'XMLHttpRequest',
  'sessionStorage',
  'localStorage',
  'Authorization',
  'Bearer ',
  'NEXT_PUBLIC_API_BASE_URL',
  '@/lib/api',
  '@/lib/auth',
] as const;

describe('pre-api boundary', () => {
  it('keeps rendered preview surfaces free from active API and browser credential markers', () => {
    const hits: string[] = [];

    for (const directory of guardedDirectories) {
      for (const file of listSourceFiles(join(process.cwd(), directory))) {
        const source = readFileSync(file, 'utf8');

        for (const marker of blockedRuntimeMarkers) {
          if (source.includes(marker)) {
            hits.push(`${file.replace(`${process.cwd()}\\`, '').replace(`${process.cwd()}/`, '')}: ${marker}`);
          }
        }
      }
    }

    expect(hits).toEqual([]);
  });
});

function listSourceFiles(directory: string): string[] {
  const entries = readdirSync(directory);
  const files: string[] = [];

  for (const entry of entries) {
    const path = join(directory, entry);
    const stat = statSync(path);

    if (stat.isDirectory()) {
      files.push(...listSourceFiles(path));
      continue;
    }

    if (/\.(ts|tsx)$/.test(entry) && !entry.endsWith('.test.ts') && !entry.endsWith('.test.tsx')) {
      files.push(path);
    }
  }

  return files;
}
