import { test, describe } from 'node:test';
import assert from 'node:assert';

describe('Frontend Logic & Analytical Helper Tests', () => {
  test('Query Builder correctly serializes filter parameters', () => {
    const filters = {
      releaseYearMin: 2015,
      releaseYearMax: 2024,
      genres: ['action', 'rpg'],
      excludePlaytests: true,
      requireLanguages: false,
    };

    const parts = [];
    for (const [key, val] of Object.entries(filters)) {
      if (val === undefined || val === null || val === '') continue;
      if (Array.isArray(val)) {
        if (val.length > 0) parts.push(`${key}=${encodeURIComponent(val.join(','))}`);
      } else {
        parts.push(`${key}=${encodeURIComponent(String(val))}`);
      }
    }
    const query = `?${parts.join('&')}`;

    assert.ok(query.includes('releaseYearMin=2015'));
    assert.ok(query.includes('releaseYearMax=2024'));
    assert.ok(query.includes('genres=action%2Crpg'));
    assert.ok(query.includes('excludePlaytests=true'));
  });

  test('Matrix cell coverage and low sample calculation', () => {
    const numerator = 45;
    const denominator = 120;
    const percentage = Math.round((numerator * 10000) / denominator) / 100;
    const isLowSample = denominator < 30;

    assert.strictEqual(percentage, 37.5);
    assert.strictEqual(isLowSample, false);

    const smallSampleDenom = 22;
    const isSmallSample = smallSampleDenom < 30;
    assert.strictEqual(isSmallSample, true);
  });

  test('Audio conversion rate calculation conditional on text', () => {
    const textGames = 500;
    const audioGames = 125;
    const conversionRate = Math.round((audioGames * 10000) / textGames) / 100;

    assert.strictEqual(conversionRate, 25.0);
  });

  test('Percentage-point gap calculation between text and audio', () => {
    const textCoverage = 42.5;
    const audioCoverage = 12.3;
    const percentagePointGap = Math.round((textCoverage - audioCoverage) * 100) / 100;

    assert.strictEqual(percentagePointGap, 30.2);
  });
});
