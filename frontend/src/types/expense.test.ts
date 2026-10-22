import { describe, it, expect } from 'vitest';
import { ALL_CATEGORIES, formatCategory } from './expense';

describe('expense types and helpers', () => {
  it('should have all 7 categories defined', () => {
    expect(ALL_CATEGORIES).toHaveLength(7);
    expect(ALL_CATEGORIES).toContain('еда');
    expect(ALL_CATEGORIES).toContain('транспорт');
    expect(ALL_CATEGORIES).toContain('жильё');
    expect(ALL_CATEGORIES).toContain('развлечения');
    expect(ALL_CATEGORIES).toContain('здоровье');
    expect(ALL_CATEGORIES).toContain('покупки');
    expect(ALL_CATEGORIES).toContain('другое');
  });

  it('formatCategory should capitalize first letter', () => {
    expect(formatCategory('еда')).toBe('Еда');
    expect(formatCategory('транспорт')).toBe('Транспорт');
    expect(formatCategory('РАЗВЛЕЧЕНИЯ')).toBe('Развлечения');
  });

  it('formatCategory should handle empty string or undefined', () => {
    expect(formatCategory('')).toBe('');
    expect(formatCategory(undefined)).toBe('');
  });
});
