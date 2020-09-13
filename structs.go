package main

import (
	"bufio"
    "fmt"
    "strings"
    "os"
	"log"
	"strconv"
	"archivosGO/utility"
)
 
const makedk string = "mkdisk";
const tamaño string= "-size";
const unit string = "-unit";
const path string = "-path";
const fit string= "-fit";
const rmdisk string = "rmdisk";
const fdisk string = "fdisk";
const tipo string = "-type";
const eliminar string = "-delete";
const name string = "-name";
const add string = "-add";
const montar string = "mount";
const desmontar string = "unmount";
const id_pt string = "-id";
const mkfs string ="mkfs";
const login string="login";
const usr string="-usr";
const pwd string="-pwd";
const id_login string="-id";
const reporte string ="rep";
const rutaAlternativa string ="-ruta"

const puta string= "file.txt";


func main()  {
	reader := bufio.NewReader(os.Stdin)	
	var entryPath string="";
	for true{
		entry, _ := reader.ReadString('\n')
		exec:= strings.Split(entry, " ");
		if(string(exec[0])=="exec"){
			entryPath=string(exec[1])[:len(string(exec[1]))-2]		
			break;
		}
	}
	file, err := os.Open(entryPath)
    if err != nil {
		fmt.Println(err)
		log.Fatal(err)	
    }
	scanner := bufio.NewScanner(file);
	var textline string;
	var concatenate bool;
	for scanner.Scan(){
		line :=scanner.Text();
		if strings.Contains(line, "\\*"){
			textline = strings.ToLower(strings.ReplaceAll(line,"\\*"," "));
			concatenate = true;
			continue;
		}
		if (concatenate == true){
			textline += strings.ToLower(line);
			fmt.Println(textline);	
			if(strings.Contains(textline,"#")){
				continue;
			}else if(strings.Contains(textline,"pause")){
				pausa, _ := reader.ReadString('\n')
				fmt.Println(pausa);
			}
			interprete(textline); 
			concatenate= false;
			continue;
		}
		textline=strings.ToLower(line);
		fmt.Println(textline);	
		if(strings.Contains(textline,"#")){
			continue;
		}else if(strings.Contains(textline,"pause")){
			pausa, _ := reader.ReadString('\n')
			fmt.Println(pausa);
		}
		interprete(textline); 
	}
}

func interprete(line string){
	var cCoutes bool= false;
	var jumpI int = 0;
	parametros:= strings.Split(line, " ");
	size:=len(parametros);
	if strings.Contains(string(parametros[0]),makedk){ //checking for the comand make disk
		var ruta string = " ";
		var nombre string = " ";
		var unidad byte ='m';
		var dksize int64 =0;
		for i:= 1; i<size;i++ {
			if strings.Contains(string(parametros[i]), "\""){
				for true{
					parametros[i]+=" "+parametros[i+1+jumpI];
					if strings.Contains(string(parametros[i+1+jumpI]), "\""){	
						jumpI +=1;	
						break;
					}
					jumpI +=1;
				}
				cCoutes= true;
			}
			parametro:=strings.Split(string(parametros[i]), "->");
			if strings.Contains(string(parametro[0]),name){
				nombre = string(parametro[1])
			}
			if strings.Contains(string(parametro[0]),path){
				ruta = string(parametro[1]);
			}
			if strings.Contains(string(parametro[0]),tamaño){
				x:=string(parametro[1]);
				t,err := strconv.ParseInt(x, 10, 64);
				dksize = t;
				if err != nil {
					log.Fatal(err)
				}
			}
			if strings.Contains(string(parametro[0]),unit){
				unidad = parametro[1][0];
			}
			if cCoutes == true{
				i+=jumpI;
				jumpI =0;
				if i>= size{
					break;
				}   
			}
		}
		utility.CreateDk(nombre,ruta,unidad,dksize);
		return;
	}	
	if strings.Contains(string(parametros[0]),rmdisk){	
		var ruta string = " ";
		for i:= 1; i<size;i++ {
			if strings.Contains(string(parametros[i]), "\""){
				for true{
					parametros[i]+=" "+parametros[i+1+jumpI];
					if strings.Contains(string(parametros[i+1+jumpI]), "\""){	
						jumpI +=1;	
						break;
					}
					jumpI +=1;
				}
				cCoutes= true;
			}
			parametro:=strings.Split(string(parametros[i]), "->");
			if strings.Contains(string(parametro[0]),path){
				ruta = string(parametro[1]);
			}
			if cCoutes == true{
				i+=jumpI;
				jumpI =0;
				if i>= size{
					break;
				}  
			}
		}
		dir := strings.ReplaceAll(ruta,"\"","");
		utility.DeleteDisk(dir);
		return;
	}	
	if strings.Contains(string(parametros[0]),fdisk){
		var ruta string = " ";
		var nombre string = " ";
		var unidad byte ='k';
		var ptsize int64 = 0;
		var addU int64 = 0;
		var delete string = " ";
		var pttipo byte = 'p';
		var ptfit byte = 'w';
		for i:= 1; i<size;i++ {
			if strings.Contains(string(parametros[i]), "\""){
				for true{
					parametros[i]+=" "+parametros[i+1+jumpI];
					if strings.Contains(string(parametros[i+1+jumpI]), "\""){	
						jumpI +=1;	
						break;
					}
					jumpI +=1;
				}
				cCoutes= true;
			}
			parametro:=strings.Split(string(parametros[i]), "->");
			if strings.Contains(string(parametro[0]),name){
				nombre = string(parametro[1])
			}
			if strings.Contains(string(parametro[0]),path){
				ruta = string(parametro[1]);
			}
			if strings.Contains(string(parametro[0]),tamaño){
				x:=string(parametro[1]);
				t,err := strconv.ParseInt(x, 10, 64);	
				ptsize=t;		
				if err != nil {
					log.Fatal(err)
				}
			}
			if strings.Contains(string(parametro[0]),unit){
				unidad = parametro[1][0];
			}
			if strings.Contains(string(parametro[0]),tipo){
				pttipo = parametro[1][0];
			}
			if strings.Contains(string(parametro[0]),fit){
				ptfit = parametro[1][0];
			}
			if strings.Contains(string(parametro[0]),add){
				x:=string(parametro[1]);
				t,err := strconv.ParseInt(x, 10, 64);	
				addU = t;		
				if err != nil {
					log.Fatal(err)
				}
			}
			if strings.Contains(string(parametro[0]),eliminar){
				delete = string(parametro[1])
			}
			if cCoutes == true{
				i+=jumpI;
				jumpI =0;
				if i>= size{
					break;
				}   
			}
		}
		nombrewC := strings.ReplaceAll(nombre,"\"","");
		utility.CreatePartition(nombrewC ,ruta ,ptfit ,ptsize ,addU ,unidad ,delete,pttipo);
		return;
	}
	if(len(parametros)==1){
		utility.PrintMP();
		return;
	}
	if strings.Contains(string(parametros[0]),desmontar){	
		var id string = " ";
		for i:= 1; i<size;i++ {
			if strings.Contains(string(parametros[i]), "\""){
				for true{
					parametros[i]+=" "+parametros[i+1+jumpI];
					if strings.Contains(string(parametros[i+1+jumpI]), "\""){	
						jumpI +=1;	
						break;
					}
					jumpI +=1;
				}
				cCoutes= true;
			}
			parametro:=strings.Split(string(parametros[i]), "->");
			Stmp:=repNums(string(parametro[0]))
			if strings.Contains(Stmp,id_pt){				
				id = string(parametro[1]);
				utility.UnMountPT(id);
			}
			if cCoutes == true{
				i+=jumpI;
				jumpI =0;
				if i>= size{
					break;
				}  
			}
		}		
		return;
	}	
	if strings.Contains(string(parametros[0]),montar){	
		var ruta string = " ";
		var nombre string = " ";
		for i:= 1; i<size;i++ {
			if strings.Contains(string(parametros[i]), "\""){
				for true{
					parametros[i]+=" "+parametros[i+1+jumpI];
					if strings.Contains(string(parametros[i+1+jumpI]), "\""){	
						jumpI +=1;	
						break;
					}
					jumpI +=1;
				}
				cCoutes= true;
			}
			parametro:=strings.Split(string(parametros[i]), "->");
			if strings.Contains(string(parametro[0]),path){
				ruta = string(parametro[1]);
			}
			if strings.Contains(string(parametro[0]),name){
				nombre = string(parametro[1])
			}
			if cCoutes == true{
				i+=jumpI;
				jumpI =0;
				if i>= size{
					break;
				}  
			}
		}
		dir := strings.ReplaceAll(ruta,"\"","");
		utility.MountPT(nombre,dir);
		return;
	}	
	//analizador del sistema de archivos
	if strings.Contains(string(parametros[0]),reporte){	
		var id string = " ";
		var rep_path string = " ";
		var t_reporte string =" ";
		var rep_Apath string =" ";
		for i:= 1; i<size;i++ {
			if strings.Contains(string(parametros[i]), "\""){
				for true{
					parametros[i]+=" "+parametros[i+1+jumpI];
					if strings.Contains(string(parametros[i+1+jumpI]), "\""){	
						jumpI +=1;	
						break;
					}
					jumpI +=1;
				}
				cCoutes= true;
			}
			parametro:=strings.Split(string(parametros[i]), "->");
			if strings.Contains(string(parametro[0]),id_pt){
				id = string(parametro[1]);
			}
			if strings.Contains(string(parametro[0]),path){
				rep_path = string(parametro[1]);
			}
			if strings.Contains(string(parametro[0]),rutaAlternativa){
				rep_Apath = string(parametro[1]);
			}
			if strings.Contains(string(parametro[0]),name){
				t_reporte = string(parametro[1]);
			}
			if cCoutes == true{
				i+=jumpI;
				jumpI =0;
				if i>= size{
					break;
				}  
			}
		}
		//llamada al metodo mkfs
		fmt.Println("  "+t_reporte+"  "+rep_Apath);
		typeOfReport(t_reporte,id,rep_path)
		return;
	}	
	if strings.Contains(string(parametros[0]),mkfs){	
		var id string = " ";
		var mtipo string = " ";
		var addU int64 = 0;
		var unidad byte ='k';
		for i:= 1; i<size;i++ {
			if strings.Contains(string(parametros[i]), "\""){
				for true{
					parametros[i]+=" "+parametros[i+1+jumpI];
					if strings.Contains(string(parametros[i+1+jumpI]), "\""){	
						jumpI +=1;	
						break;
					}
					jumpI +=1;
				}
				cCoutes= true;
			}
			parametro:=strings.Split(string(parametros[i]), "->");
			if strings.Contains(string(parametro[0]),id_pt){
				id = string(parametro[1]);
			}
			if strings.Contains(string(parametro[0]),tipo){
				mtipo = string(parametro[1])
			}
			if strings.Contains(string(parametro[0]),add){
				x:=string(parametro[1]);
				t,err := strconv.ParseInt(x, 10, 64);	
				addU = t;		
				if err != nil {
					log.Fatal(err)
				}
			}
			if strings.Contains(string(parametro[0]),unit){
				unidad = parametro[1][0];
			}
			if cCoutes == true{
				i+=jumpI;
				jumpI =0;
				if i>= size{
					break;
				}  
			}
		}
		//llamada al metodo mkfs
		fmt.Printf("id: %s tipo: %s unidades: %d unidad: %c\n",id,mtipo,addU,unidad);
		return;
	}	
	if strings.Contains(string(parametros[0]),login){	
		var usuario string = " ";
		var password string = " ";
		var id_particion string = " ";
		for i:= 1; i<size;i++ {
			if strings.Contains(string(parametros[i]), "\""){
				for true{
					parametros[i]+=" "+parametros[i+1+jumpI];
					if strings.Contains(string(parametros[i+1+jumpI]), "\""){	
						jumpI +=1;	
						break;
					}
					jumpI +=1;
				}
				cCoutes= true;
			}
			parametro:=strings.Split(string(parametros[i]), "->");
			if strings.Contains(string(parametro[0]),id_login){
				id_particion = string(parametro[1]);
			}
			if strings.Contains(string(parametro[0]),usr){
				usuario = string(parametro[1])
			}
			if strings.Contains(string(parametro[0]),pwd){
				password = string(parametro[1])
			}
			if cCoutes == true{
				i+=jumpI;
				jumpI =0;
				if i>= size{
					break;
				}  
			}
		}
		//llamada al metodo login
		fmt.Printf("usuario: %s contraseña: %s id: %s \n",usuario,password,id_particion);
		return;
	}	
}

func typeOfReport(t_reporte string ,id string,rep_path string){
	if(t_reporte=="mbr"){
		utility.ReporteMBR(id,rep_path);
		return;
	}
	if(t_reporte=="disk"){
		utility.ReporteDISK(id,rep_path);
		return;
	}
}

func repNums(s string) string {
	out := make([]rune, len(s)) 

	i, added := 0, false
	for _, r := range s {
		if r >= '0' && r <= '9' {
			if added {
				continue
			}
			added, out[i] = true, r
		} else {
			added, out[i] = false, r
		}
		i++
	}
	return string(out[:i])
}



