import { Card, CardContent, CardHeader } from '@/shared/ui/primitives/card'
import { Skeleton } from '@/shared/ui/primitives/skeleton'
import { SkeletonRows, Th } from '../../../shared/ui'

export function LaporanSkeleton() {
  return (
    <div className="space-y-5" aria-busy="true" aria-label="Memuat laporan">
      <div className="grid grid-cols-2 gap-4 xl:grid-cols-4">
        {[0, 1, 2, 3].map((i) => (
          <Card key={i} className="opc-kpi">
            <div className="opc-kpi-body">
              <Skeleton className="size-16 shrink-0 rounded-md" />
              <div className="min-w-0 space-y-2">
                <Skeleton className="h-7 w-28" />
                <Skeleton className="h-4 w-24" />
                <Skeleton className="h-3 w-32" />
              </div>
            </div>
          </Card>
        ))}
      </div>

      <div className="grid gap-4 lg:grid-cols-[2fr_1fr]">
        <Card>
          <CardHeader>
            <Skeleton className="h-5 w-36" />
            <Skeleton className="h-4 w-48" />
          </CardHeader>
          <CardContent>
            <div className="flex h-64 items-end gap-3" aria-hidden="true">
              {[38, 62, 48, 78, 55, 70, 44, 66].map((h, i) => (
                <Skeleton key={i} className="flex-1 rounded-b-none rounded-t-md" style={{ height: `${h}%` }} />
              ))}
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <Skeleton className="h-5 w-40" />
            <Skeleton className="h-4 w-24" />
          </CardHeader>
          <CardContent>
            <Skeleton className="mx-auto size-36 rounded-full" />
            <div className="mt-4 space-y-2">
              {[0, 1, 2].map((i) => (
                <div key={i} className="flex items-center justify-between gap-3">
                  <span className="flex items-center gap-2">
                    <Skeleton className="size-2.5 rounded-full" />
                    <Skeleton className="h-4 w-20" />
                  </span>
                  <Skeleton className="h-4 w-16" />
                </div>
              ))}
            </div>
          </CardContent>
        </Card>
      </div>

      <div className="grid gap-4 lg:grid-cols-[1fr_1.4fr]">
        <Card>
          <CardHeader>
            <Skeleton className="h-5 w-36" />
            <Skeleton className="h-4 w-28" />
          </CardHeader>
          <CardContent>
            <div className="space-y-3" aria-hidden="true">
              {[0, 1, 2].map((i) => (
                <div key={i} className="flex items-center justify-between gap-3">
                  <Skeleton className="h-5 w-16 rounded-full" />
                  <Skeleton className="h-4 w-8" />
                </div>
              ))}
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <Skeleton className="h-5 w-40" />
            <Skeleton className="h-4 w-44" />
          </CardHeader>
          <CardContent>
            <table className="w-full border-collapse">
              <thead>
                <tr><Th>Tanggal</Th><Th>ID</Th><Th>Kasir</Th><Th>Metode</Th><Th right>Total</Th></tr>
              </thead>
              <SkeletonRows cols={5} rows={5} />
            </table>
          </CardContent>
        </Card>
      </div>
    </div>
  )
}
