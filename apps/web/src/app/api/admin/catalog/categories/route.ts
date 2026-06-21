import { NextRequest } from 'next/server';

import { handleAdminCatalogCategories } from '../../../../../lib/server/catalog';

export async function GET(request: NextRequest) {
  return handleAdminCatalogCategories(request);
}
