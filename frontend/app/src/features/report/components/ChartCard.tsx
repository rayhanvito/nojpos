import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/shared/ui/primitives/card'

export function ChartCard({ title, sub, action, children }: { title: string; sub: string; action?: React.ReactNode; children: React.ReactNode }) {
  return (
    <Card>
      <CardHeader className="flex flex-row items-start justify-between gap-3 space-y-0">
        <div className="space-y-1.5">
          <CardTitle>{title}</CardTitle>
          <CardDescription>{sub}</CardDescription>
        </div>
        {action}
      </CardHeader>
      <CardContent>{children}</CardContent>
    </Card>
  )
}
