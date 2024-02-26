package disco

import (
	"github.com/bytesparadise/libasciidoc/pkg/types"
	"github.com/hasty/alchemy/ascii"
	"github.com/hasty/alchemy/matter"
	"github.com/hasty/alchemy/matter/constraint"
)

func fixConstraintCells(doc *ascii.Doc, rows []*types.TableRow, columnMap ascii.ColumnIndex) (err error) {
	if len(rows) < 2 {
		return
	}
	constraintIndex, ok := columnMap[matter.TableColumnConstraint]
	if !ok {
		return
	}
	for _, row := range rows[1:] {
		cell := row.Cells[constraintIndex]
		vc, e := ascii.RenderTableCell(cell)
		if e != nil {
			continue
		}

		dataType := doc.ReadRowDataType(row, columnMap, matter.TableColumnType)
		if dataType != nil {
			c := constraint.ParseString(vc)
			fixed := c.AsciiDocString(dataType)
			if fixed != vc {
				err = setCellString(cell, fixed)
				if err != nil {
					return
				}
			}
		}

	}
	return
}
