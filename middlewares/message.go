package middlewares

var Responses = struct {
	FailedValidations      *NewRM
	InternalServerError    *NewRM
	UserNotFound           *NewRM
	InvalidRoles           *NewRM
	TimeFieldRequired      *NewRM
	EndTimeBeforeStartTime *NewRM
	EndTimeBeforeNow       *NewRM
	EventNotFound          *NewRM
}{
	FailedValidations: &NewRM{
		Language.English: "Failed field validations",
		Language.Spanish: "Las validaciones de los campos fallaron",
	},
	InternalServerError: &NewRM{
		Language.English: "Internal server error",
		Language.Spanish: "Problemas con el servidor",
	},

	UserNotFound: &NewRM{
		Language.English: "User not found",
		Language.Spanish: "No se encontró el usuario",
	},
	InvalidRoles: &NewRM{
		Language.English: "Invalid roles",
		Language.Spanish: "No tienes permiso para realizar esta acción",
	},
	TimeFieldRequired: &NewRM{
		Language.English: "Time field is required",
		Language.Spanish: "El campo de hora es requerido",
	},
	EndTimeBeforeStartTime: &NewRM{
		Language.English: "End time can't be before start time",
		Language.Spanish: "La hora de término no puede ser antes de la de inicio",
	},
	EndTimeBeforeNow: &NewRM{
		Language.English: "End time can't be before now",
		Language.Spanish: "La hora de término no puede ser antes de la actual",
	},
	EventNotFound: &NewRM{
		Language.English: "EventNotFound",
		Language.Spanish: "El evento no existe",
	},
}

type NewRM map[string]string

var Language = struct {
	English string
	Spanish string
}{
	English: "en",
	Spanish: "es",
}

var LanguageMap = map[string]string{
	Language.Spanish: "Spanish",
	Language.English: "English",
}
