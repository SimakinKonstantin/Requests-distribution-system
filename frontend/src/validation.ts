export const MAX_TEXT_LENGTH = 50
export const MAX_LENGTH_ERROR_MESSAGE = `Значение не должно превышать ${MAX_TEXT_LENGTH} символов`
export const MIN_LENGTH_ERROR_MESSAGE = 'Значение не может состоять из одного символа'
export const FIRST_CHAR_INVALID_MESSAGE = 'Недопустимый первый символ'

export const PERSON_NAME_PATTERN = '^[А-Яа-яЁё]+(?:[ -][А-Яа-яЁё]+)*$'
export const SURNAME_PATTERN = '^[А-Яа-яЁё]+(?:-[А-Яа-яЁё]+)*$'
export const RUSSIAN_TEXT_PATTERN = "^[А-Яа-яЁё\\s.,!?():;\"'\\-]+$"

const personNameRegex = new RegExp(PERSON_NAME_PATTERN)
const surnameRegex = new RegExp(SURNAME_PATTERN)
const russianTextRegex = new RegExp(RUSSIAN_TEXT_PATTERN)

function startsWithInvalidFirstChar(value: string): boolean {
  return /^[ЪъЬь]/.test(value)
}

export function applyTextFieldValidity(form: HTMLFormElement) {
  const fields = form.querySelectorAll<HTMLInputElement | HTMLTextAreaElement>(
    'input[data-text-field="true"], textarea[data-text-field="true"]',
  )

  fields.forEach(field => {
    if (!field.dataset.validityBound) {
      field.addEventListener('input', () => field.setCustomValidity(''))
      field.dataset.validityBound = 'true'
    }
    const normalized = field.value.trim()
    field.setCustomValidity('')
    if (startsWithInvalidFirstChar(normalized)) {
      field.setCustomValidity(FIRST_CHAR_INVALID_MESSAGE)
      return
    }
    if (normalized.length < 2) {
      field.setCustomValidity(MIN_LENGTH_ERROR_MESSAGE)
      return
    }
    if (normalized.length > MAX_TEXT_LENGTH) {
      field.setCustomValidity(MAX_LENGTH_ERROR_MESSAGE)
    }
  })
}

function validateLength(value: string, label: string, min = 2): string | null {
  const normalized = value.trim()
  if (startsWithInvalidFirstChar(normalized)) return FIRST_CHAR_INVALID_MESSAGE
  if (normalized.length < min) return min === 2 ? MIN_LENGTH_ERROR_MESSAGE : `${label}: минимум ${min} символа`
  if (normalized.length > MAX_TEXT_LENGTH) return `${label}: максимум ${MAX_TEXT_LENGTH} символов`
  return null
}

export function validatePersonName(value: string, label: string): string | null {
  const lengthError = validateLength(value, label, 2)
  if (lengthError) return lengthError
  if (!personNameRegex.test(value.trim())) {
    return `${label}: только русские буквы, пробел и дефис (без спецсимволов по краям)`
  }
  return null
}

export function validateSurname(value: string): string | null {
  const lengthError = validateLength(value, 'Фамилия', 2)
  if (lengthError) return lengthError
  if (!surnameRegex.test(value.trim())) {
    return 'Фамилия: только русские буквы и дефис, дефис не может быть первым или последним'
  }
  return null
}

export function validateEntityName(value: string, label: string): string | null {
  return validatePersonName(value, label)
}

export function validateAppealText(value: string, label = 'Текст'): string | null {
  const normalized = value.trim()
  if (normalized.length === 0) return null
  if (startsWithInvalidFirstChar(normalized)) return FIRST_CHAR_INVALID_MESSAGE
  if (!russianTextRegex.test(normalized)) {
    return `${label}: используйте русские буквы и базовые знаки пунктуации`
  }
  return null
}

export function validateEmail(value: string): string | null {
  const normalized = value.trim()
  if (normalized.length === 0) return 'Email: поле обязательно'
  if (startsWithInvalidFirstChar(normalized)) return FIRST_CHAR_INVALID_MESSAGE
  if (normalized.length < 2) return MIN_LENGTH_ERROR_MESSAGE
  if (normalized.length > MAX_TEXT_LENGTH) return `Email: максимум ${MAX_TEXT_LENGTH} символов`
  const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/
  if (!emailRegex.test(normalized)) return 'Email: некорректный формат'
  return null
}

export function validateNonNegativeNumber(value: number, label: string): string | null {
  if (!Number.isFinite(value)) return `${label}: некорректное число`
  if (value < 0) return `${label}: значение должно быть неотрицательным`
  return null
}
