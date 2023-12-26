package constraint

import (
	"fmt"

	"github.com/hasty/alchemy/matter"
)

type MinConstraint struct {
	Minimum matter.ConstraintLimit `json:"min"`
}

func (c *MinConstraint) Type() matter.ConstraintType {
	return matter.ConstraintTypeMin
}

func (c *MinConstraint) AsciiDocString(dataType *matter.DataType) string {
	return fmt.Sprintf("min %s", c.Minimum.AsciiDocString(dataType))
}

func (c *MinConstraint) Equal(o matter.Constraint) bool {
	if oc, ok := o.(*MinConstraint); ok {
		return oc.Minimum.Equal(c.Minimum)
	}
	return false
}

func (c *MinConstraint) Min(cc *matter.ConstraintContext) (min matter.DataTypeExtreme) {
	return c.Minimum.Min(cc)
}

func (c *MinConstraint) Max(cc *matter.ConstraintContext) (max matter.DataTypeExtreme) {
	return
}

func (c *MinConstraint) Default(cc *matter.ConstraintContext) (max matter.DataTypeExtreme) {
	return
}

func (c *MinConstraint) Clone() matter.Constraint {
	return &MinConstraint{Minimum: c.Minimum.Clone()}
}
