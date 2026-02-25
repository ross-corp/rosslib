import { render, screen } from '@testing-library/react'
import { expect, test, vi } from 'vitest'
import React from 'react'
import BookList from './book-list'

// Mock next/link
vi.mock('next/link', () => {
  return {
    default: ({ children, href }: { children: React.ReactNode; href: string }) => (
      <a href={href}>{children}</a>
    ),
  }
})

// Mock StatusPicker
vi.mock('@/components/shelf-picker', () => {
  return {
    default: () => <div data-testid="status-picker">Status Picker</div>,
  }
})

// Mock StarRating to simplify test
vi.mock('@/components/star-rating', () => {
  return {
    default: ({ rating }: { rating: number }) => (
      <div data-testid="star-rating">{rating}</div>
    ),
  }
})

test('renders book list correctly', () => {
  const books = [
    {
      key: '/works/OL123W',
      title: 'Test Book',
      authors: ['Test Author'],
      publish_year: 2023,
      cover_url: 'http://example.com/cover.jpg',
      average_rating: 4.5,
      rating_count: 10,
      already_read_count: 5,
    },
  ]

  render(
    <BookList
      books={books}
      statusValues={[]}
      statusKeyId="key"
      bookStatusMap={{}}
    />
  )

  expect(screen.getByText('Test Book')).toBeDefined()
  expect(screen.getByText('Test Author')).toBeDefined()
  expect(screen.getByText('2023')).toBeDefined()
  expect(screen.getByText('5 reads')).toBeDefined()

  // Check if StatusPicker is rendered (via mock)
  expect(screen.getByTestId('status-picker')).toBeDefined()

  // Check if StarRating is rendered (via mock)
  expect(screen.getByTestId('star-rating')).toBeDefined()
  expect(screen.getByText('4.5')).toBeDefined()
})

test('renders nothing if no books', () => {
  const { container } = render(
    <BookList
      books={[]}
      statusValues={[]}
      statusKeyId="key"
      bookStatusMap={{}}
    />
  )
  expect(container.firstChild).toBeNull()
})
