import { expect, it } from 'vitest';
import { recordsFromPayload } from './records';

it('uses backend items as table rows without calculating values', () => {
  expect(recordsFromPayload({ items: [{ product_name: 'Kopi', on_hand_quantity: 8 }] })).toEqual([{ product_name: 'Kopi', on_hand_quantity: 8 }]);
});
