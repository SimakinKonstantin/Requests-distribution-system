import { fireEvent, render, screen } from '@testing-library/react'
import { describe, expect, it, vi } from 'vitest'
import Modal from './Modal'

describe('Modal', () => {
  it('renders title and content', () => {
    render(
      <Modal title="Тестовый заголовок" onClose={vi.fn()}>
        <div>Содержимое</div>
      </Modal>,
    )

    expect(screen.getByText('Тестовый заголовок')).toBeInTheDocument()
    expect(screen.getByText('Содержимое')).toBeInTheDocument()
  })

  it('calls onClose when close button clicked', () => {
    const onClose = vi.fn()
    render(
      <Modal title="Modal" onClose={onClose}>
        <span>Body</span>
      </Modal>,
    )

    fireEvent.click(screen.getByRole('button'))
    expect(onClose).toHaveBeenCalledTimes(1)
  })
})

