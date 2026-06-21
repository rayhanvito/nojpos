import { Badge } from '@/components/ui/badge';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';

type DisabledField = {
  label: string;
  value: string;
  helper?: string;
};

type DisabledFormPreviewProps = {
  title: string;
  description: string;
  fields: readonly DisabledField[];
};

export function DisabledFormPreview({ title, description, fields }: DisabledFormPreviewProps) {
  return (
    <Card className="disabled-form-card">
      <CardHeader>
        <div>
          <span className="card-kicker">Form contoh terkunci</span>
          <CardTitle>{title}</CardTitle>
        </div>
        <Badge variant="warning">Dikunci</Badge>
      </CardHeader>
      <CardContent>
        <CardDescription>{description}</CardDescription>
        <div className="disabled-form-grid">
          {fields.map((field) => (
            <label key={field.label}>
              <span>{field.label}</span>
              <input value={field.value} readOnly disabled aria-label={field.label} />
              {field.helper ? <small>{field.helper}</small> : null}
            </label>
          ))}
        </div>
      </CardContent>
    </Card>
  );
}
