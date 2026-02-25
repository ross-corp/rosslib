import { render, screen } from '@testing-library/react'
import { expect, test } from 'vitest'
import React from 'react'
import StarRating from './star-rating'

test('renders star rating correctly', () => {
  render(<StarRating rating={3.5} count={10} />)

  // 3.5 rounds to 4 filled stars
  const starsContainer = screen.getByText(/⭐⭐⭐⭐☆/i)
  expect(starsContainer).toBeDefined()

  // Check if rating number is displayed
  expect(screen.getByText('3.50')).toBeDefined()

  // Check if count is displayed
  expect(screen.getByText('(10)')).toBeDefined()
})

test('renders empty stars correctly', () => {
  render(<StarRating rating={0} />)

  const starsContainer = screen.getByText(/☆☆☆☆☆/i)
  expect(starsContainer).toBeDefined()

  expect(screen.getByText('0.00')).toBeDefined()
})
