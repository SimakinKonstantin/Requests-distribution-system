import { describe, expect, it } from 'vitest'
import {
  FIRST_CHAR_INVALID_MESSAGE,
  MIN_LENGTH_ERROR_MESSAGE,
  validateAppealText,
  validateEmail,
  validateEntityName,
  validateNonNegativeNumber,
  validatePersonName,
  validateSurname,
} from './validation'

describe('validation', () => {
  it('validates person name and entity name', () => {
    const testCases = [
      { fn: validatePersonName, value: 'Иван', label: 'Имя', want: null },
      { fn: validatePersonName, value: 'Ъван', label: 'Имя', want: FIRST_CHAR_INVALID_MESSAGE },
      { fn: validatePersonName, value: 'A', label: 'Имя', want: MIN_LENGTH_ERROR_MESSAGE },
      { fn: validateEntityName, value: 'Команда-1', label: 'Название', wantIncludes: 'только русские буквы' },
    ]

    testCases.forEach(tc => {
      const result = tc.fn(tc.value, tc.label)
      if (tc.wantIncludes) {
        expect(result).toContain(tc.wantIncludes)
        return
      }
      expect(result).toBe(tc.want)
    })
  })

  it('validates surname', () => {
    const testCases = [
      { value: 'Петров', want: null },
      { value: '-Петров', wantIncludes: 'только русские буквы' },
    ]

    testCases.forEach(tc => {
      const result = validateSurname(tc.value)
      if (tc.wantIncludes) {
        expect(result).toContain(tc.wantIncludes)
        return
      }
      expect(result).toBe(tc.want)
    })
  })

  it('validates appeal text', () => {
    const testCases = [
      { value: '', want: null },
      { value: 'Тест обращения', want: null },
      { value: 'Ъпример', want: FIRST_CHAR_INVALID_MESSAGE },
      { value: 'Hello', wantIncludes: 'используйте русские буквы' },
    ]

    testCases.forEach(tc => {
      const result = validateAppealText(tc.value)
      if (tc.wantIncludes) {
        expect(result).toContain(tc.wantIncludes)
        return
      }
      expect(result).toBe(tc.want)
    })
  })

  it('validates email and number', () => {
    const emailCases = [
      { value: '', wantIncludes: 'поле обязательно' },
      { value: 'a', want: MIN_LENGTH_ERROR_MESSAGE },
      { value: 'test@example.com', want: null },
      { value: 'bad@@mail', wantIncludes: 'некорректный формат' },
    ]

    emailCases.forEach(tc => {
      const result = validateEmail(tc.value)
      if (tc.wantIncludes) {
        expect(result).toContain(tc.wantIncludes)
        return
      }
      expect(result).toBe(tc.want)
    })

    const numberCases = [
      { value: 0, want: null },
      { value: 10, want: null },
      { value: -1, wantIncludes: 'неотрицательным' },
      { value: Number.NaN, wantIncludes: 'некорректное число' },
    ]

    numberCases.forEach(tc => {
      const result = validateNonNegativeNumber(tc.value, 'Лимит')
      if (tc.wantIncludes) {
        expect(result).toContain(tc.wantIncludes)
        return
      }
      expect(result).toBe(tc.want)
    })
  })
})

