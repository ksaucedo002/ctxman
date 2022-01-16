## CTXMAN
Es un helper que nos ayuda a realizar consultas dinámicas usando GORM, en cuanto al límite de registros consultados y el offset, además podemos indicar que campos pueden omitirse, lo cuales no se consultaran a base de datos. 

* Fuciona con: `no esta soportado por otros web  frameworks`
    * GORM go orm
    * ECHO go web framework

Ejemplo
```go
    func (h *handler) FindAll(c echo.Context) error {
        data, err := h.service.FindAll(ctxman.Newctxx(c)) ///ctxman prepara y recupera parametros en el context
        if err != nil {
            log.Error(err)
            return c.String(http.StatusBadRequest, "algo pasow")
        }
        return c.JSON(http.StatusOK, data)
    }

    // Implementación de capa logica o services
    type Service struct {
	    repo repositries.Repository
    }   

    // Implementación de la interfaz Omiter de CTXMAN
    // la cual ayuda a indicar que atributos son omitibles
    // y cuáles tienen que ser pre cargados por GORM
    func (s *Service) OmitFiels() ([]string, []string) { 
        // Primer slice indica atributos omitibles y el segundo los pre cargados
        /// Gorm precargara todo los campos que aparescan en el slide       
	    return []string{"Nombre"}, []string{"Libros"}
    }
    func (s *Service) FindAll(ctx ctxman.Ctxx) ([]*models.Data, error) {
        /// Preparamos el contexto pasándole la implementación de la interfaz Omiter
        return s.repo.FindAll(ctx.WithOmiter(s))
    }
    func (s *editorialService) FindByID(ctx ctxman.Ctxx, ID uint) (*models.Data, error) {
        /// Preparamos el contexto pasándole la implementación de la interfaz Omiter
	    return s.r.FindByCode(ctx.WithOmiter(s), ID)
    }

    //Implementación Repository

    type Repository struct {
	    grom_conn *gorm.DB
    }
    func (r *Repository) FindAll(ctx ctxman.Ctxx) ([]*models.Data, error) {
        datos := []*models.Data{}
        tx := ctx.FormatGORM(r.grom_conn) /// Configuramos la conexion de gorm
        if err := tx.Find(&datos).Error; err != nil {
            return nil, err
        }
        return datos, nil
    }
    func (r *Repository) FindByID(ctx ctxman.Ctxx, code uint) (*models.Data, error) {
        data := models.Data{}
        // SimpleGORM no tiene en cuenta limit y offset
        tx := ctx.SimpleGORM(r.grom_conn) /// Configuramos la conexion de gorm
        if err := tx.Find(&data, "id=?", code).Error; err != nil {
            return nil, err
        }
        return &data, nil
    }
```
Al consultar un url podremos pasar en los query params los campos:
- `limit`: indica el numero de registros
- `offset`: indica desde donde ser leera los registros, o desplazamiento
- `omit`:  Los campos que se desean omitir en la consulta, estos deben de estar separados por comas y tal y como se definieron en la interfaz omiter, probablemente CamelCase

exameple:
    - GET: `http://localhost:9091/datos?Omit=Description,Address,Monto&offset=10&limit=10`