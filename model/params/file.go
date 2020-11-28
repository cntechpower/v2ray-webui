package params

type ReadFileParam struct {
	FileName string `form:"file_name" binding:"required"`
	From     int64  `form:"from"`
	To       int64  `form:"to" binding:"required"`
}
