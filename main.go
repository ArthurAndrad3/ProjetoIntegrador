package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	_ "github.com/lib/pq"
)

const (
	host     = "monorail.proxy.rlwy.net"
	port     = 19413
	user     = "postgres"
	password = "apdDfNTgAlMRpNuRbtUHzXKuQoUSbizr"
	dbname   = "railway"
)

var db *sql.DB

func init() {
	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)

	var err error
	db, err = sql.Open("postgres", psqlInfo)
	if err != nil {
		log.Fatal(err)
	}

	err = db.Ping()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Conexão com o banco de dados estabelecida com sucesso!")
}

type Usuario struct {
	Email    string `json:"email"`
	CPF      string `json:"cpf"`
	Telefone string `json:"telefone"`
	Username string `json:"username"`
	Password string `json:"password"`
}

type Paciente struct {
	Nome               string `json:"nome"`
	CPF                string `json:"cpf"`
	DataNascimento     string `json:"data_nascimento"`
	Sexo               string `json:"sexo"`
	Telefone           string `json:"telefone"`
	Email              string `json:"email"`
	NomeMae            string `json:"nome_mae"`
	CEP                string `json:"cep"`
	Estado             string `json:"estado"`
	Cidade             string `json:"cidade"`
	Endereco           string `json:"endereco"`
	HomemMaiorQuarenta bool   `json:"homem_maior_quarenta"`
	Etilista           bool   `json:"etilista"`
	LesaoSuspeita      bool   `json:"lesao_suspeita"`
	Tabagista          bool   `json:"tabagista"`
	DataCadastro       string `json:"data_cadastro"`
	Microarea          string `json:"microarea"`
	Encaminhado        bool   `json:"encaminhado"`
}

func enableCORS(w http.ResponseWriter) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
}

func CadastrarPacientes(w http.ResponseWriter, r *http.Request) {
	enableCORS(w)
	if r.Method != http.MethodPost {
		http.Error(w, "Método não permitido", http.StatusMethodNotAllowed)
		return
	}

	var paciente Paciente
	err := json.NewDecoder(r.Body).Decode(&paciente)
	if err != nil {
		http.Error(w, "Erro ao processar o JSON: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Define o campo Encaminhado como true
	paciente.Encaminhado = true

	sqlStatement := `
	INSERT INTO pacientes (nomepaciente, cpf, nascimento, sexo, telefone, email, nomemae, cep, estado, cidade, endereco, "40anos", etilista, lesao, tabagista, cadastro, microarea, encaminhado)
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18)`

	_, err = db.Exec(sqlStatement, paciente.Nome, paciente.CPF, paciente.DataNascimento, paciente.Sexo, paciente.Telefone, paciente.Email, paciente.NomeMae, paciente.CEP, paciente.Estado, paciente.Cidade, paciente.Endereco, paciente.HomemMaiorQuarenta, paciente.Etilista, paciente.LesaoSuspeita, paciente.Tabagista, paciente.DataCadastro, paciente.Microarea, paciente.Encaminhado)
	if err != nil {
		http.Error(w, fmt.Sprintf("Erro ao inserir dados no banco de dados: %v", err), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"message": "Paciente cadastrado com sucesso"})
}
func ListarNomesEDatasPacientes(w http.ResponseWriter, r *http.Request) {
	enableCORS(w)

	log.Println("Iniciando ListarNomesEDatasPacientes")

	rows, err := db.Query("SELECT nomepaciente, nascimento FROM pacientes")
	if err != nil {
		log.Printf("Erro ao executar consulta SQL: %v", err)
		http.Error(w, fmt.Sprintf("Erro ao buscar nomes e datas dos pacientes: %v", err), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	type NomeDataPaciente struct {
		Nome           string `json:"nome"`
		DataNascimento string `json:"data_nascimento"`
	}

	var pacientes []NomeDataPaciente
	for rows.Next() {
		var paciente NomeDataPaciente
		err := rows.Scan(&paciente.Nome, &paciente.DataNascimento)
		if err != nil {
			log.Printf("Erro ao ler dados do paciente: %v", err)
			http.Error(w, fmt.Sprintf("Erro ao ler dados do paciente: %v", err), http.StatusInternalServerError)
			return
		}
		pacientes = append(pacientes, paciente)
	}

	if err := rows.Err(); err != nil {
		log.Printf("Erro nos resultados do banco de dados: %v", err)
		http.Error(w, fmt.Sprintf("Erro nos resultados do banco de dados: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(pacientes)
	if err != nil {
		log.Printf("Erro ao codificar resposta JSON: %v", err)
		http.Error(w, fmt.Sprintf("Erro ao codificar resposta JSON: %v", err), http.StatusInternalServerError)
		return
	}

	log.Printf("Dados retornados com sucesso: %v", pacientes)
}
func ListarPacientesEncaminhados(w http.ResponseWriter, r *http.Request) {
	enableCORS(w)

	rows, err := db.Query("SELECT nomepaciente FROM pacientes WHERE encaminhado = true")
	if err != nil {
		http.Error(w, fmt.Sprintf("Erro ao buscar pacientes encaminhados: %v", err), http.StatusInternalServerError)
		return
	}
	defer rows.Close()
	type PacienteEncaminhado struct {
		Nome string `json:"nome"`
	}
	var pacientes []PacienteEncaminhado // Assumindo que você tenha uma struct Paciente definida

	for rows.Next() {
		var paciente PacienteEncaminhado // Estrutura para armazenar dados do paciente
		err := rows.Scan(&paciente.Nome) // Nome como uma propriedade da struct Paciente
		if err != nil {
			http.Error(w, fmt.Sprintf("Erro ao ler nome do paciente: %v", err), http.StatusInternalServerError)
			return
		}
		pacientes = append(pacientes, paciente)
	}

	if err := rows.Err(); err != nil {
		http.Error(w, fmt.Sprintf("Erro nos resultados do banco de dados: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(pacientes)
}
func ListarPacientesAbsenteistas(w http.ResponseWriter, r *http.Request) {
	enableCORS(w)

	rows, err := db.Query("SELECT nomepaciente FROM pacientes WHERE encaminhado = false")
	if err != nil {
		http.Error(w, fmt.Sprintf("Erro ao buscar pacientes encaminhados: %v", err), http.StatusInternalServerError)
		return
	}
	defer rows.Close()
	type PacienteAbsenteistas struct {
		Nome string `json:"nome"`
	}
	var pacientes []PacienteAbsenteistas // Assumindo que você tenha uma struct Paciente definida

	for rows.Next() {
		var paciente PacienteAbsenteistas // Estrutura para armazenar dados do paciente
		err := rows.Scan(&paciente.Nome)  // Nome como uma propriedade da struct Paciente
		if err != nil {
			http.Error(w, fmt.Sprintf("Erro ao ler nome do paciente: %v", err), http.StatusInternalServerError)
			return
		}
		pacientes = append(pacientes, paciente)
	}

	if err := rows.Err(); err != nil {
		http.Error(w, fmt.Sprintf("Erro nos resultados do banco de dados: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(pacientes)
}

func main() {

	http.HandleFunc("/cadastrar-paciente", CadastrarPacientes)
	http.HandleFunc("/listar-nomes-datas-pacientes", ListarNomesEDatasPacientes)
	http.HandleFunc("/listar-pacientes-encaminhados", ListarPacientesEncaminhados)
	http.HandleFunc("/listar-pacientes-absenteistas", ListarPacientesAbsenteistas)

	fmt.Println("Servidor iniciado na porta 8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
