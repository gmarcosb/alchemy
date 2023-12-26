package constraint

import "github.com/hasty/alchemy/matter"

type DescribedConstraint struct {
}

func (c *DescribedConstraint) Type() matter.ConstraintType {
	return matter.ConstraintTypeDescribed
}

func (c *DescribedConstraint) AsciiDocString(dataType *matter.DataType) string {
	return "desc"
}

func (c *DescribedConstraint) Equal(o matter.Constraint) bool {
	_, ok := o.(*DescribedConstraint)
	return ok
}

func (c *DescribedConstraint) Min(cc *matter.ConstraintContext) (min matter.DataTypeExtreme) {
	return
}

func (c *DescribedConstraint) Max(cc *matter.ConstraintContext) (max matter.DataTypeExtreme) {
	return
}

func (c *DescribedConstraint) Default(cc *matter.ConstraintContext) (max matter.DataTypeExtreme) {
	return
}

func (c *DescribedConstraint) Clone() matter.Constraint {
	return &DescribedConstraint{}
}
