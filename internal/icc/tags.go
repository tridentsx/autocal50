package icc

// Tag signatures for display profiles (ICC.1:2022 §9).
var (
	TagProfileDesc    = sig("desc")
	TagCopyright      = sig("cprt")
	TagMediaWhitePoint = sig("wtpt")
	TagChromAdapt     = sig("chad")
	TagRedMatrixCol   = sig("rXYZ")
	TagGreenMatrixCol = sig("gXYZ")
	TagBlueMatrixCol  = sig("bXYZ")
	TagRedTRC         = sig("rTRC")
	TagGreenTRC       = sig("gTRC")
	TagBlueTRC        = sig("bTRC")
)

// Tag type signatures (ICC.1:2022 §10).
var (
	TypeXYZ              = sig("XYZ ")
	TypeCurve            = sig("curv")
	TypeParametricCurve  = sig("para")
	TypeTextDesc         = sig("desc")
	TypeMLUC             = sig("mluc")
	TypeS15Fixed16Array  = sig("sf32")
)
