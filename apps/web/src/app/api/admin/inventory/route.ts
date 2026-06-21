import { NextRequest } from 'next/server';

import { handleAdminInventory } from '../../../../lib/server/inventory';

export async function GET(request: NextRequest) {
  return handleAdminInventory(request);
}
