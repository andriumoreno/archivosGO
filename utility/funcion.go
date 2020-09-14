package utility

import (
    "fmt"
	"time"
	"os"
	"strings"
	"path"
	"math/rand"
	"encoding/binary"
	//"unsafe"
	//"bufio"
	"io"
	"strconv"
	"os/exec"
	"bytes"
)

const mbrPS int64 = 38;
const mbrPT int64 = 39;
const mbrPF int64 = 40;
const mbrPST int64 = 41;
const mbrPSZ int64 = 49;
const mbrPN int64 = 57;
const ebrst int64 = 178;
const sizeOfMBR int64 = 200;
const sizeOfEBR int64 = 48;
const ebrnext int64 = 196;//

const tr0 string="<TR><TD>";
const tr1 string="</TD><TD><TABLE BORDER=\"0\"><TR><TD>";
const tr2 string="</TD></TR></TABLE></TD></TR>"
const Cabecera string="<TR><TD>NOMBRE</TD><TD><TABLE BORDER=\"0\"><TR><TD>VALOR</TD></TR></TABLE></TD></TR>"


var mount_pt mount;

type mount_Data struct {
    pt_name string;
    pt_id string;
}

type mount_Dk struct {
    dk_path string;
    data []mount_Data
}

type mount struct {
    disco []mount_Dk;
}

type EPartition struct{
    Part_status byte ;
    Part_fit byte;
    Part_start int64 ;
	Part_size int64 ;
	Part_next int64 ;
    Part_name[16]byte;
}

type List_l struct{
	Lista []EPartition;
}

type Partition struct{
    Part_status byte ;
    Part_type byte;
    Part_fit byte;
    Part_start uint64 ;
    Part_size uint64 ;
    Part_name[16]byte;
}

type MBR struct{
	Mbr_tamano uint64;
    Mbr_fecha_creacion [22]byte;
    Mbr_disk_signature uint64;
    Mbr_partition[4]Partition;
}

func getTime() string {
	currentTime := time.Now()
	date:=currentTime.Format("2006-01-02 3:4:5")

	return date;
}

func CheckBR(err error){
	if err != nil {
		fmt.Println("binary.Read failed:", err)
	 }
}

func check(e error) {
    if e != nil {
       fmt.Println("revisar parametros");
    }
}

func makeDir(path string){
	parametros:= strings.Split(path, "/");
	size:=len(parametros); 
	var texto string = "";
	for i:=0;i<size;i++ {
		texto=texto+"/"+parametros[i];
		existDir(texto);		
	}
}

func existDir(path string){	
	if _, err := os.Stat(path); err == nil {
	return;
  	} else if os.IsNotExist(err) {
		err := os.Mkdir(path, 0755)
		check(err)
  	} else {
  	fmt.Println("error. ruta no creada");  
 	}
}

func CreateDk(name string, ruta string, unit byte, disksize int64){
	dir := strings.ReplaceAll(ruta,"\"","");
	makeDir(dir);
 	var x int64= 0;
    if unit == 'k'{
        x = disksize * 1024 - 1;
    } else if unit=='m'{
        x = disksize * 1024 * 1024 - 1;
	}else{
		x =disksize
	}
	disk, err := os.Create(path.Join("/"+dir, name));
	check(err);
	disk.Truncate(int64(x));
	disk.Seek(0,0);	 
	rand.Seed(time.Now().UnixNano());
	hd := MBR{Mbr_tamano:uint64(x), Mbr_disk_signature: uint64(rand.Intn(200))}  
	copy(hd.Mbr_fecha_creacion[:],getTime());
	err = binary.Write(disk, binary.LittleEndian, hd)  
	CheckBR(err);	
	disk.Close();
}

func DeleteDisk(path string){
	e := os.Remove(path) 
    if e != nil { 
		fmt.Println("El archivo no existe, verifique la ruta");
		return;
	} 
	fmt.Println("Archivo eliminado exitosamente");
}

/*
at this point, the program will write the partition in the file in the file.
first it will check if there is a partition avaliable, then it will check the partition's name doesn't exist
at last it will verify there is enough space for the particion.

before to jump to create the particion it will check if the opcional comands are empty. in this case the opcional 
parameters are add y delete. also it has to check what type of partition it will write in the file
*/

func CreatePartition(name string, ruta string, fit byte, ptsize int64, add int64, unit byte, delete string, pttype byte){
	dir := strings.ReplaceAll(ruta,"\"","");
	disk, err := os.OpenFile("/"+dir, os.O_RDWR ,0666);
	defer disk.Close();
	check(err);
	var x int64 = 0;
    if unit == 'k'{
        x = ptsize * 1024;
    } else if unit == 'm'{
        x = ptsize * 1024 * 1024;
	} else {
		x = ptsize
	}
	if(delete!=" "){
		deletePartition(disk,name,delete);
		return;
	}
	if(add!=0){
		addBytes(disk,name,add)
		return;
	}
	//create a primary partition or an extended
	offset:=checkPTAvaliable(disk, mbrPS);
	if(offset==0&&pttype!='l'){
		fmt.Println("the four partitions are unavailable");
		return;
	}
	var cast[16]byte;
	copy(cast[:],name);
	Sname:=checkName(disk, mbrPN,cast);
	notL:= nullLP(disk);
	if(notL==true){
		Lname:=checkNamePL(disk,196,cast);
		if(Lname==true){
			fmt.Println("there is a logical partition with this name");
			return;
		}
	}
	if(Sname==true){
		fmt.Println("there is a partition with this name");
		return;
	}
	e_partition:=searchE(disk, mbrPT);
	if(e_partition==true&&pttype =='e'){
		fmt.Println("there only can be one extended partition per disc");
		return
	}	
	if(pttype =='l'){
		var ebr EPartition;
		ebr.Part_status='1'
		ebr.Part_fit=fit;
		ebr.Part_size=x;
		ebr.Part_next = -1;
		copy(ebr.Part_name[:],cast[:]);
		CreateLP(disk,ebr,ptsize)	
		return;
	}
	Espace:=checkSpace(offset,x,disk);
	if(Espace==0){
		fmt.Println("there is no enough space to create the partition");
		return;
	}
	//writing the values of the partition on the file
	disk.Seek(offset,0);
	a := make([]byte, 8)
	binary.LittleEndian.PutUint64(a, uint64(Espace))
	b := make([]byte, 8)
	binary.LittleEndian.PutUint64(b, uint64(ptsize))
	status:=[]byte{'1',pttype,fit};
	bytearray:=append(status,a...);
	bytearray= append(bytearray,b...);
	slice := cast[:];
	bytearray= append(bytearray,slice...);
	err =binary.Write(disk, binary.LittleEndian, &bytearray)  
	CheckBR(err);
}

func CreateLP(file *os.File, lp EPartition,size_l int64){
	eP:=searchE(file,mbrPT);
	if(eP == true){
		file.Seek(2, io.SeekCurrent);
		var E_start int64;
		binary.Read(file, binary.LittleEndian, &E_start)
		file.Seek(0, io.SeekCurrent);
		var E_size int64;
		binary.Read(file, binary.LittleEndian, &E_size)
		no:=nullLP(file);
		noLP:= findNextL(file,222);
		if(no==false&&noLP==0){
			if(E_size>=sizeOfEBR+size_l){
				lp.Part_start = E_start +sizeOfEBR;
				file.Seek(ebrst,0);
				err:=binary.Write(file, binary.LittleEndian, &lp)
				CheckBR(err);
				return;				
			}
			fmt.Println("there is not enough space to create a logical partition")
			return;
		}else if(no==false){
			file.Seek(196,0);
			var t_next int64;
			binary.Read(file, binary.LittleEndian, &t_next)
			if(t_next!=0){
				lp.Part_start = E_start +sizeOfEBR;
				if(size_l<=t_next-lp.Part_start){
					lp.Part_start = E_start +sizeOfEBR;
					file.Seek(ebrst,0);
					err:=binary.Write(file, binary.LittleEndian, &lp)
					CheckBR(err);
					return;			
				}
				fmt.Println("there is not enough space to create a logical partition")
				return;
			}else{
				lp.Part_start = E_start +sizeOfEBR;
				if(size_l<=noLP-lp.Part_start){
					lp.Part_start = E_start +sizeOfEBR;
					file.Seek(ebrst,0);
					err:=binary.Write(file, binary.LittleEndian, &lp)
					CheckBR(err);
					return;			
				}
				fmt.Println("there is not enough space to create a logical partition")
				return;
			}
		}
		offset:=LPTAvaliable(file,178);
		noLP= findNextL(file,offset+44);
		if(noLP!=0){
			readInt:= offset -40;
			var L_start int64;
			file.Seek(readInt, 0);
			binary.Read(file, binary.LittleEndian, &L_start)
			var L_size int64;
			binary.Read(file, binary.LittleEndian, &L_size)
			changest:=L_size+L_start;
			file.Seek(0, io.SeekCurrent);
			binary.Write(file, binary.LittleEndian, &changest)
			lp.Part_start = changest;  
			if(size_l<=noLP-lp.Part_start){
				file.Seek(offset,0);
				lp.Part_next=noLP;
				err:=binary.Write(file, binary.LittleEndian, &lp)
				CheckBR(err);
				return;
			}
			fmt.Println("there is not enough space to create a logical partition")
			return;
		}else{
			if(E_size-spaceL(file, 196)>=size_l){
				offset:=checkLPTAvaliable(file,196);
				readInt:= offset -40;
				var L_start int64;
				file.Seek(readInt, 0);
				binary.Read(file, binary.LittleEndian, &L_start)
				var L_size int64;
				binary.Read(file, binary.LittleEndian, &L_size)
				changest:=L_size+L_start;
				file.Seek(0, io.SeekCurrent);
				binary.Write(file, binary.LittleEndian, &changest)
				file.Seek(offset,0);
				lp.Part_start = changest;  
				err:=binary.Write(file, binary.LittleEndian, &lp)
				CheckBR(err);
				return;
			}
			fmt.Println("there is not enough space to create a logical partition")
			return;
		}
	}
	fmt.Println("this disk has no extended partition");
	return;
}

//
func nullLP(file *os.File)bool{
	file.Seek(ebrst,0);
	var status byte;
	err := binary.Read(file, binary.LittleEndian, &status)
	CheckBR(err);
	if(status!=0){
		return true;
	}
	return false;
}

//
func checkNamePL(file *os.File,offsetnt int64 ,nombre[16]byte)bool{
	file.Seek(offsetnt,0);
	var next int64;
	binary.Read(file, binary.LittleEndian, &next)
	file.Seek(0, io.SeekCurrent)
	var buffer[16]byte;
	binary.Read(file, binary.LittleEndian, &buffer)
	if(nombre == buffer){
		return true;
	}
	if(next!=-1){
		return checkNamePL(file,offsetnt+42,nombre);
	}
	return false;
}

func spaceL(file *os.File,offsetnt int64)int64{
	file.Seek(offsetnt,0);
	var next int64;
	binary.Read(file, binary.LittleEndian, &next)
	file.Seek(-16, io.SeekCurrent)
	var buffer int64;
	binary.Read(file, binary.LittleEndian, &buffer)
	buffer = buffer + sizeOfEBR;
	if(next!=-1){
		buffer = buffer +spaceL(file,offsetnt+42);
	}
	return buffer;
}

/*
the function check name y check space work diferent depending on the type of the partition
*/

func checkName(file *os.File,ptstart int64, nombre[16]byte)bool{
	if(ptstart>=178){
		return false;
	}
	file.Seek(ptstart,0);
	var buffer[16]byte;
	err := binary.Read(file, binary.LittleEndian, &buffer)
	CheckBR(err);
	if (nombre == buffer){
		return true;
	}
	return checkName(file,ptstart+35,nombre);
}
//
func searchE(file *os.File,ptstart int64)bool{
	if(ptstart>=178){
		return false;
	}
	file.Seek(ptstart,0);
	var buffer byte;
	err := binary.Read(file, binary.LittleEndian, &buffer)
	auxint:= ptstart -1;
	file.Seek(auxint,0);
	var status byte;
	err = binary.Read(file, binary.LittleEndian, &status)
	CheckBR(err);
	if ('e' == buffer && '1'==status){
		return true;
	}
	return searchE(file,ptstart+35);
}

/*
this function will verify there is enough space to create the partition 
and there is 4 cases to check
	CASE 1:
	the function of this case is create a partition in the first position. 
	1) when the disk is empty
	2) when there is an existing partition
	CASE 2:
	the fuction of this case is create a partition in the second position 
	1) when there is an active partition at the first position and the rest of the disk is empty
	2) when there are two active partition, in this situacion the first and the third or quarter partition
*/

func checkSpace(offset int64, size int64, file *os.File)int64{
	var dksize int64;
	var byte_start int64;
	file.Seek(0,0);
	err := binary.Read(file, binary.LittleEndian, &dksize)
	CheckBR(err);
	next_pt:= diskEmpty(file, offset); 
	if(next_pt!=0){
		file.Seek(next_pt+3,0);			
		err := binary.Read(file, binary.LittleEndian, &byte_start)
		CheckBR(err);
	}
	var space int64 = 0;
	switch offset {	
    case 38: 	
		if(next_pt==0){
			avaliable_space:= dksize-sizeOfMBR+1;
			if(size<=avaliable_space){
				space = sizeOfMBR;
			}
		}else{			
			avaliable_space:= byte_start-sizeOfMBR;
			if(size<=avaliable_space){
				space = sizeOfMBR;
			}
		}
	case 73:
		var fp_size int64 = 0 ;
		file.Seek(mbrPSZ,0);
		binary.Read(file, binary.LittleEndian, &fp_size);
		var fp_byte_start int64 = 0 ;
		file.Seek(mbrPST,0);
		binary.Read(file, binary.LittleEndian, &fp_byte_start);
		if(next_pt==0){
			avaliable_space:= dksize-sizeOfMBR-fp_size+1;
			if(size<=avaliable_space){
				space = fp_byte_start+fp_size;
			}
		}else{			
			avaliable_space:= byte_start-sizeOfMBR-fp_size;
			if(size<=avaliable_space){
				space = fp_byte_start+fp_size;
			}
		}
    case 108:
		var fp_size int64 = 0 ;
		var sp_size int64 = 0;
		file.Seek(mbrPSZ,0);
		binary.Read(file, binary.LittleEndian, &fp_size);
		file.Seek(mbrPSZ+35,0);
		binary.Read(file, binary.LittleEndian, &sp_size);
		var sp_byte_start int64 = 0 ;
		file.Seek(mbrPST+35,0);
		binary.Read(file, binary.LittleEndian, &sp_byte_start);
		if(next_pt==0){
			avaliable_space:= dksize-sizeOfMBR-fp_size-sp_size+1;
			if(size<=avaliable_space){
				space = sp_byte_start+sp_size;
			}
		}else{			
			avaliable_space:= byte_start-sizeOfMBR-fp_size-sp_size;
			if(size<=avaliable_space){
				space = sp_byte_start+sp_size;
			}
		}
	case 143:
		var fp_size int64 = 0;
		var sp_size int64 = 0;
		var tp_size int64 = 0;
		file.Seek(mbrPSZ,0);
		binary.Read(file, binary.LittleEndian, &fp_size);
		file.Seek(mbrPSZ+35,0);
		binary.Read(file, binary.LittleEndian, &sp_size);
		file.Seek(mbrPSZ+70,0);
		binary.Read(file, binary.LittleEndian, &tp_size);
		var tp_byte_start int64 = 0 ;
		file.Seek(mbrPST+70,0);
		binary.Read(file, binary.LittleEndian, &tp_byte_start);
		avaliable_space:= dksize-sizeOfMBR-fp_size-sp_size-tp_size+1;
		if(size<=avaliable_space){
			space = tp_byte_start+tp_size;
    	}
	}
	return space;
}
//
func checkPTAvaliable(file *os.File,ptstart int64)int64{
	if(ptstart>=178){
		return 0;
	}
	file.Seek(ptstart,0);
	var status byte = ' ';
	err := binary.Read(file, binary.LittleEndian, &status)
	CheckBR(err);
	if(status==0){
		return ptstart;
	}
	if(status!='0'){
		return checkPTAvaliable(file,ptstart+35);
	}
	return ptstart;
}

//
func checkLPTAvaliable(file *os.File,ptstart int64)int64{
	file.Seek(ptstart,0);
	var next int64;
	binary.Read(file, binary.LittleEndian, &next)
	if(next!=-1){
		return checkLPTAvaliable(file,ptstart+42);
	}
	return ptstart+24;
}

func LPTAvaliable(file *os.File, offset int64)int64{
	file.Seek(offset,0);
	var next byte;
	binary.Read(file, binary.LittleEndian, &next)
	if(next!=0){
		return LPTAvaliable(file,offset+42);
	}
	return offset;
}

/* 
return 0 if the disk is empty or return the offset position where is the next active partition
*/

func diskEmpty(file *os.File,ptstart int64)int64{
	if(ptstart>=178){
		return 0;
	}
	file.Seek(ptstart,0);
	var status byte = ' ';
	err := binary.Read(file, binary.LittleEndian, &status)
	CheckBR(err);
	if(status!='1'){
		return diskEmpty(file,ptstart+35);
	}
	return ptstart;
}

func MountPT(name string, ruta string){
	dir := strings.ReplaceAll(ruta,"\"","");
	disk, err := os.OpenFile("/"+dir, os.O_RDWR ,0666);
	defer disk.Close();
	check(err);
	var cast[16]byte;
	copy(cast[:],name);
	exist:=checkName(disk, mbrPN, cast);
	notL:= nullLP(disk);
	if(notL==true){
		Lname:=checkNamePL(disk,196,cast);
		if(Lname==true){
			exist= true;
		}
	}
	if(checkMount(name)==true){
		fmt.Println("ya hay una particion montada con este nombre");
		return;
	}
	if(exist == true){
		if(len(mount_pt.disco)==0){			
			var tmp mount_Dk;
			var tmp1 mount_Data; 
			tmp.dk_path = ruta;
			tmp1.pt_name = name;
			tmp1.pt_id = "vda1";
			tmp.data = append(tmp.data,tmp1);
			mount_pt.disco=append(mount_pt.disco,tmp);
		}else{
			for i := range mount_pt.disco {
				if(mount_pt.disco[i].dk_path==ruta){
					var tmp mount_Data; 
					x:=len(mount_pt.disco[i].data);
					tmp.pt_name = name;
					stringaux:= "vd"+string(toChar(i+1))+strconv.Itoa(x+1);
					auxbool:= checkId(stringaux);
					if(auxbool==true){
						stringaux= "vd"+string(toChar(i))+strconv.Itoa(x);
						auxbool1:= checkId(stringaux);
						if(auxbool1==false){
							tmp.pt_id =stringaux;
						}
						stringaux= "vd"+string(toChar(i+2))+strconv.Itoa(x+2);
						auxbool2:= checkId(stringaux);
						if(auxbool2==false){
							tmp.pt_id =stringaux;
						}
					}else{
					tmp.pt_id =stringaux;
					} 
					mount_pt.disco[i].data= append(mount_pt.disco[i].data,tmp);
					return;
				}			
			}
			var tmp mount_Dk;
			var tmp1 mount_Data; 
			tmp.dk_path = ruta;
			tmp1.pt_name = name;
			tmp1.pt_id = "vd"+string(toChar(len(mount_pt.disco)+1))+"1";
			tmp.data = append(tmp.data,tmp1);
			mount_pt.disco=append(mount_pt.disco,tmp);
			return;
		}
	}else{
		fmt.Println("check the name of the partition");
	}
}

func PrintMP(){
	if(len(mount_pt.disco)==0){
		fmt.Println("there are not mount partitions");
	}
	for i := range mount_pt.disco {
		for j := range mount_pt.disco[i].data{
			fmt.Println("id->"+mount_pt.disco[i].data[j].pt_id+" -path-> "+ mount_pt.disco[i].dk_path+" -name-> "+mount_pt.disco[i].data[j].pt_name);
		}	
	}
}

func checkMount(name string)bool{
	if(len(mount_pt.disco)==0){
		fmt.Println("there are not mount partitions");
		return false;
	}
	for i := range mount_pt.disco {
		for j := range mount_pt.disco[i].data{
			if(name==mount_pt.disco[i].data[j].pt_name){
				return true;
			}
		}	
	}
	return false;
}

func checkId(id string)bool{
	if(len(mount_pt.disco)==0){
		fmt.Println("there are not mount partitions");
		return false;
	}
	for i := range mount_pt.disco {
		for j := range mount_pt.disco[i].data{
			if(id==mount_pt.disco[i].data[j].pt_id){
				return true;
			}
		}	
	}
	return false;
}

func ReturnPath(id string)string{
	if(len(mount_pt.disco)==0){
		fmt.Println("there are not mount partitions");
		return " ";
	}
	for i := range mount_pt.disco {
		for j := range mount_pt.disco[i].data{
			if(id==mount_pt.disco[i].data[j].pt_id){
				return mount_pt.disco[i].dk_path;
			}
		}	
	}
	return " ";
}

func toChar(i int) rune {
    return rune('a' - 1 + i)
}

func UnMountPT(id string){
	for i := range mount_pt.disco {
		if(len(mount_pt.disco[i].data)==0){
			mount_pt.disco=removeDk(mount_pt.disco,i)
			fmt.Println("the disk has no more partitions");
			return;
		}
		for j := range mount_pt.disco[i].data{
			if(mount_pt.disco[i].data[j].pt_id==id){
				mount_pt.disco[i].data=removeData(mount_pt.disco[i].data,j);
				fmt.Println("the partition has been remove");
				return;
			}			
		}	
	}
}

func removeData(slice []mount_Data, s int) []mount_Data {
    return append(slice[:s], slice[s+1:]...)
}

func removeDk(slice []mount_Dk, s int) []mount_Dk {
    return append(slice[:s], slice[s+1:]...)
}

func deletePartition(file *os.File,name string, parametro string){
	var p_errase[35] byte;
	var l_errase[42] byte; 
	var cero byte;
	cero =0;
	var cast[16]byte;
	copy(cast[:],name);
	if(parametro=="fast"){
		exist:=checkName(file,mbrPN,cast);
		if(exist==true){
			file.Seek( -35, io.SeekCurrent );		
			er:=binary.Write(file, binary.LittleEndian, &cero)  
			CheckBR(er);   
			file.Seek( 0, io.SeekCurrent );
			var pt_type byte;
			err := binary.Read(file, binary.LittleEndian, &pt_type)
			CheckBR(err);
			fmt.Println(file.Seek( 0, io.SeekCurrent ));
			fmt.Println(pt_type)
			if(pt_type=='e'){
				fi, err := file.Stat()
				if err != nil {
				  fmt.Println("error al eliminar particion extendida");
				}	
				delete_L:=fi.Size()-180;
				errase_l := make([]byte, delete_L);
				file.Seek( ebrst, 0 );
				err = binary.Write(file, binary.LittleEndian, &errase_l)  
				CheckBR(err);
				fmt.Println("particion extendida eliminada exitosamente");
				return;
			}		
			return;
		}
		notL:= nullLP(file);
		if(notL==true){
			Lname:=checkNamePL(file,196,cast);
			if(Lname==true){
				deleteOffset,fatal:=file.Seek( -42, io.SeekCurrent );
				if fatal != nil {
					fmt.Println(fatal);
				}
				var auxE EPartition;
				k:=binary.Read(file, binary.LittleEndian, &auxE);
				if k != nil {
					fmt.Println(k);
				}
				if(deleteOffset!=178){
					for i:=deleteOffset-42; i>=178; i=i-42 {
						file.Seek(i,0);
						var status_d byte;
						binary.Read(file, binary.LittleEndian, &status_d);
						if(status_d=='1'){
							file.Seek( 17, io.SeekCurrent );
							erroB:=binary.Write(file, binary.LittleEndian, &auxE.Part_next) 
							CheckBR(erroB); 
							break;
						}
					}
				}
				file.Seek(deleteOffset,0)
				er:=binary.Write(file, binary.LittleEndian, &cero)  
				CheckBR(er);
				fmt.Println("particion logica eliminada con exito"); 
				return;
			}
		}		
		fmt.Println("there is no a partition with this name: "+name);	
	}else if(parametro=="full"){
		exist:=checkName(file,mbrPN,cast);
		if(exist==true){
			file.Seek( -34, io.SeekCurrent );	
			var pt_type byte;
			err := binary.Read(file, binary.LittleEndian, &pt_type)
			CheckBR(err);	  
			file.Seek( -2, io.SeekCurrent );
			er:=binary.Write(file, binary.LittleEndian, &p_errase)  
			CheckBR(er); 
			if(pt_type=='e'){
				fi, err := file.Stat()
				if err != nil {
				  fmt.Println("error al eliminar particion extendida");
				}	
				delete_L:=fi.Size()-180;
				errase_l := make([]byte, delete_L);
				file.Seek( ebrst, 0 );
				err = binary.Write(file, binary.LittleEndian, &errase_l)  
				CheckBR(err);
				fmt.Println("particion extendida eliminada exitosamente");
			}		
			return;
		}
		notL:= nullLP(file);
		if(notL==true){
			Lname:=checkNamePL(file,196,cast);
			if(Lname==true){
				fmt.Println("deleting a logical partition");
				deleteOffset,fatal:=file.Seek( -42, io.SeekCurrent );
				if fatal != nil {
					fmt.Println(fatal);
				}
				var auxE EPartition;
				k:=binary.Read(file, binary.LittleEndian, &auxE);
				if k != nil {
					fmt.Println(k);
				}
				if(deleteOffset!=178){
					for i:=deleteOffset-42; i>=178; i=i-42 {
						file.Seek(i,0);
						var status_d byte;
						binary.Read(file, binary.LittleEndian, &status_d);
						if(status_d=='1'){
							file.Seek( 17, io.SeekCurrent );
							erroB:=binary.Write(file, binary.LittleEndian, &auxE.Part_next);
							CheckBR(erroB); 
							break;
						}
					}
				}
				file.Seek(deleteOffset,0)
				er:=binary.Write(file, binary.LittleEndian, &l_errase)  
				CheckBR(er);
				fmt.Println("particion logica eliminada con exito"); 
				return;
			}
		}		
		fmt.Println("there is no a partition with this name: "+name);	
	}

}

func findNextL(file *os.File,offset int64)int64{
	var intaux int64=offset;
	var b_start int64=0;
	for i := 0; i < 10; i++ {
		file.Seek(intaux,0);		
		binary.Read(file, binary.LittleEndian, &b_start)
		if(b_start!=0){
			return b_start;
		}
		intaux= intaux+42;
	}
	return 0;
}

func addBytes(file *os.File,name string,byteToAdd int64){
	var varaux int64=0;
	var cast[16]byte;
	copy(cast[:],name);
	var hasNext byte=0;
	var byte_start int64=0;
	primaria:=checkName(file,mbrPN,cast);
	if(primaria==true){
		finalP, i:=file.Seek(0,io.SeekCurrent);
		if i != nil {
			fmt.Println("error al agregar bytes");
		}
		binary.Read(file, binary.LittleEndian, &hasNext)
		file.Seek(-1,io.SeekCurrent)
		if(hasNext==0&&finalP<178){
			for true{
				where, j:=file.Seek(35,io.SeekCurrent);
				if j != nil {
					fmt.Println("error al agregar bytes");
				}
				binary.Read(file, binary.LittleEndian, &hasNext)
				if(where>=178){
					break;
				}
				if(hasNext!=0){
					varaux= where -finalP;
					break;
				}
			}
		}
		file.Seek(finalP+1,0);
		file.Seek(-33,io.SeekCurrent);
		binary.Read(file, binary.LittleEndian, &byte_start)
		if(hasNext==0||finalP==178){
			sizeToChange, e := file.Seek(0,io.SeekCurrent);
			if e != nil {
				fmt.Println("error al agregar bytes");
			}
			fi, err := file.Stat()
			if err != nil {
				fmt.Println("error al agregar bytes a la particion");
			}
			var size_n int64=0;
			file.Seek(sizeToChange,0);
			binary.Read(file, binary.LittleEndian, &size_n)
			if(fi.Size()-byte_start+1>=size_n+byteToAdd){
				if(size_n+byteToAdd<0){
					fmt.Println("el tamaño de la particion no puede ser negativo");
					return;
				}
				file.Seek(sizeToChange,0);
				intTemp:=size_n+byteToAdd;
				binary.Write(file, binary.LittleEndian, &intTemp)
				return;
			}
			fmt.Println("espacio insuficiente, no se pueden agregar tamaño");
			return;
		}else{
			sizeToChange, e := file.Seek(0,io.SeekCurrent);
			if e != nil {
				fmt.Println("error al agregar bytes");
			}
			file.Seek(27+varaux,io.SeekCurrent);
			var b_startN int64=0;
			var size_n int64=0;
			binary.Read(file, binary.LittleEndian, &b_startN)
			file.Seek(sizeToChange,0);
			binary.Read(file, binary.LittleEndian, &size_n)
			if(b_startN-byte_start>=size_n+byteToAdd){
				if(size_n+byteToAdd<0){
					fmt.Println("el tamaño de la particion no puede ser negativo");
					return;
				}
				file.Seek(sizeToChange,0);
				intTemp:=size_n+byteToAdd;
				binary.Write(file, binary.LittleEndian, &intTemp)
				return;
			}
			fmt.Println("espacio insuficiente, no se pueden agregar tamaño");
			return;
		}
	}
	logica:=checkNamePL(file,196,cast);
	if(logica==true){
		file.Seek(-40,io.SeekCurrent);
		binary.Read(file, binary.LittleEndian, &byte_start)
		sizeToChange, e := file.Seek(0,io.SeekCurrent);
		if e != nil {
			fmt.Println("error al agregar bytes");
		}
		var b_startN int64=0;
		var size_n int64=0;
		binary.Read(file, binary.LittleEndian, &size_n)
		binary.Read(file, binary.LittleEndian, &b_startN)
		if(b_startN==-1){
			searchE(file,mbrPT);
			file.Seek(2, io.SeekCurrent);
			var E_start int64;
			binary.Read(file, binary.LittleEndian, &E_start)
			file.Seek(0, io.SeekCurrent);
			var E_size int64;
			binary.Read(file, binary.LittleEndian, &E_size)
			if(E_size+E_start-byte_start>=size_n+byteToAdd){
				if(size_n+byteToAdd<0){
					fmt.Println("el tamaño de la particion no puede ser negativo");
					return;
				}
				file.Seek(sizeToChange,0);
				intTemp:=size_n+byteToAdd;
				binary.Write(file, binary.LittleEndian, &intTemp)
				return;
			}
			fmt.Println("espacio insuficiente, no se pueden agregar tamaño");
			return;
		}else{
			if(b_startN-byte_start>=size_n+byteToAdd){
				if(size_n+byteToAdd<0){
					fmt.Println("el tamaño de la particion no puede ser negativo");
					return;
				}
				file.Seek(sizeToChange,0);
				intTemp:=size_n+byteToAdd;
				binary.Write(file, binary.LittleEndian, &intTemp)
				return;
			}
			fmt.Println("espacio insuficiente, no se pueden agregar tamaño");
			return;
		}	
	}
}

func ReporteMBR(id string, path string){
	dir := strings.ReplaceAll(ReturnPath(id),"\"","");
	file, err := os.OpenFile("/"+dir, os.O_RDWR ,0666);
	defer file.Close();
	check(err);
	var hdrive MBR;
	file.Seek(0,0);
	e:=binary.Read(file, binary.LittleEndian, &hdrive)
	if e != nil {
		fmt.Println(err);
	}
	var grafo string = "digraph test {\n graph [ratio=fill];\n node [label=\"\\N\", fontsize=15, shape=plaintext];\n graph [bb=\"0,0,352,154\"];\n arset [label=<\n"
	grafo=grafo+"<TABLE ALIGN=\"LEFT\">\n"+Cabecera; 
	grafo=grafo+tr0+"Mbr_tamano"+tr1+strconv.FormatUint(hdrive.Mbr_tamano+1,10)+tr2;
	grafo=grafo+tr0+"Mbr_disk_signature"+tr1+strconv.FormatUint(hdrive.Mbr_disk_signature,10)+tr2;
	n := bytes.IndexByte(hdrive.Mbr_fecha_creacion[:], 0)
	grafo=grafo+tr0+"Mbr_fecha_creacion"+tr1+string(hdrive.Mbr_fecha_creacion[:n])+tr2;
	for i := 0; i < 4; i++{
		if(hdrive. Mbr_partition[i].Part_status!='1'){
			continue;
		}
		grafo=grafo+tr0+"Part_status_"+strconv.Itoa(i)+tr1+string(hdrive. Mbr_partition[i].Part_status)+tr2;
		grafo=grafo+tr0+"Part_type_"+strconv.Itoa(i)+tr1+string(hdrive. Mbr_partition[i].Part_type)+tr2;
		grafo=grafo+tr0+"Part_fit_"+strconv.Itoa(i)+tr1+string(hdrive. Mbr_partition[i].Part_fit)+tr2;
		grafo=grafo+tr0+"Part_start_"+strconv.Itoa(i)+tr1+strconv.FormatUint(hdrive. Mbr_partition[i].Part_start,10)+tr2;
		grafo=grafo+tr0+"Part_size_"+strconv.Itoa(i)+tr1+strconv.FormatUint(hdrive. Mbr_partition[i].Part_size,10)+tr2;
		n = bytes.IndexByte(hdrive.Mbr_partition[i].Part_name[:], 0)
		grafo=grafo+tr0+"Part_name_"+strconv.Itoa(i)+tr1+string(hdrive.Mbr_partition[i].Part_name[:n])+tr2;
	}
	grafo= grafo+"</TABLE>> ];}"
	f, err := os.Create("test.txt")
    if err != nil {
        fmt.Println(err)
        return
    }
    l, err := f.WriteString(grafo)
    if err != nil {
        fmt.Println(err)
        f.Close()
        return
    }
    fmt.Println(l, "bytes written successfully")
    err = f.Close()
    if err != nil {
        fmt.Println(err)
        return
    }
	exec.Command("dot C:/Users/Asus/Desktop/test.txt -Tpng -O ").Output();
}

func ReporteDISK(id string, path string){
	dir := strings.ReplaceAll(ReturnPath(id),"\"","");
	file, err := os.OpenFile("/"+dir, os.O_RDWR ,0666);
	defer file.Close();
	check(err);
	var hdrive MBR;
	file.Seek(0,0);
	e:=binary.Read(file, binary.LittleEndian, &hdrive)
	if e != nil {
		fmt.Println(e);
	}
	var myList List_l;
	for i := 0; i < 4; i++{
		if(hdrive. Mbr_partition[i].Part_type=='e'&&hdrive. Mbr_partition[i].Part_status=='1'){		
			var auxE EPartition;
			file.Seek(ebrst,0)
			k:=binary.Read(file, binary.LittleEndian, &auxE)
			if k != nil {
				fmt.Println(k);
			}
			if(auxE.Part_next==-1){
				break;
			}else{
				myList.Lista=append(myList.Lista,auxE);
				var auxint int64=42;
				count:=0;
				for true{
					file.Seek(ebrst+auxint,0);
					k=binary.Read(file, binary.LittleEndian, &auxE)
					if k != nil {
						fmt.Println(k);
					}
					if(auxE.Part_next==-1&&auxE.Part_status=='1'){
						myList.Lista=append(myList.Lista,auxE);
						break;
					}else if(auxE.Part_next==0){
						auxint=auxint+42;
						if(count>6){
							break;
						}
						continue;
					}
					if(auxE.Part_status=='1'){
					 	myList.Lista=append(myList.Lista,auxE);
					}
					auxint=auxint+42;
				}
			}
		}
	}
	var tamañoR int64=int64(hdrive.Mbr_tamano)+1-200;
	var R_logicas string= " ";
    var grafo string = "digraph test {\n graph [ratio=fill];\n node [label=\"\\N\", fontsize=15, shape=plaintext];\n graph [bb=\"0,0,352,154\"];\n arset [label=<\n <TABLE ALIGN=\"LEFT\"><TR>"
	grafo=grafo+"<TD>MBR</TD>"
	var emptySpace int64=0;
	var checklast int=0;
	for i := 3; i >= 0; i--{
		if(hdrive. Mbr_partition[i].Part_status=='1'){
			checklast = i;
			break;
		}
	}
	for i := 0; i < 4; i++{
		if(hdrive. Mbr_partition[i].Part_type=='e'&&hdrive. Mbr_partition[i].Part_status=='1'){
			porcentaje:= Percent(int64(hdrive. Mbr_partition[i].Part_size),tamañoR);
			s := fmt.Sprintf("%.2f", porcentaje) 
			R_logicas =  "<TD><TABLE><TR><TD>"
			n := bytes.IndexByte(hdrive.Mbr_partition[i].Part_name[:], 0)
			R_logicas = R_logicas +string(hdrive.Mbr_partition[i].Part_name[:n])+" (EXTENDIDA) "+s+"%";
			R_logicas = R_logicas +"</TD></TR><TR><TD><TABLE><TR><TD>EBR</TD>"
			if(len(myList.Lista)>0){
				if(int64(hdrive. Mbr_partition[i].Part_start)+sizeOfEBR!= myList.Lista[0].Part_start||myList.Lista[0].Part_status==0){
					R_logicas = R_logicas +"<TD>LIBRE: "+"</TD><TD>EBR</TD>";
				}else{
					n = bytes.IndexByte(myList.Lista[0].Part_name[:], 0)
					R_logicas = R_logicas +"<TD>LOGICA: "+string(myList.Lista[0].Part_name[:n])+"</TD><TD>EBR</TD>"
				}
				for j := 1; j < len(myList.Lista); j++{
					if(myList.Lista[j].Part_status=='1'){
						n = bytes.IndexByte(myList.Lista[j].Part_name[:], 0)
						R_logicas = R_logicas +"<TD>LOGICA: "+string(myList.Lista[j].Part_name[:n])+"</TD><TD>EBR</TD>"
					}
					if(myList.Lista[j].Part_start+myList.Lista[j].Part_size!=myList.Lista[j].Part_next&&myList.Lista[j].Part_next!=-1||myList.Lista[j].Part_status==0){
						R_logicas = R_logicas +"<TD>LIBRE: "+"</TD><TD>EBR</TD>";
					}
				}
				if(myList.Lista[len(myList.Lista)-1].Part_start+myList.Lista[len(myList.Lista)-1].Part_size!=int64(hdrive. Mbr_partition[i].Part_size)+int64(hdrive. Mbr_partition[i].Part_start)){
					R_logicas = R_logicas +"<TD>LIBRE: "+"</TD>";
				}
			}
			R_logicas = R_logicas +"</TR></TABLE></TD></TR></TABLE></TD>"
			grafo=grafo+R_logicas;
			emptySpace=int64(hdrive. Mbr_partition[i].Part_size)+int64(hdrive. Mbr_partition[i].Part_start);
			if(i==checklast&&hdrive.Mbr_tamano!=hdrive.Mbr_partition[i].Part_start+hdrive.Mbr_partition[i].Part_size){
				porcentaje:= Percent(int64(hdrive.Mbr_tamano-(hdrive. Mbr_partition[i].Part_start+hdrive. Mbr_partition[i].Part_size)),tamañoR);
				s = fmt.Sprintf("%.2f", porcentaje) 
				grafo=grafo+"<TD>LIBRE<br/> "+s+"%"+"</TD>"
			}
			continue;
		}else if(hdrive. Mbr_partition[i].Part_status=='1'){
			porcentaje:= Percent(int64(hdrive. Mbr_partition[i].Part_size),tamañoR);
			s := fmt.Sprintf("%.2f", porcentaje) 
			if(i>0&&emptySpace!=int64(hdrive. Mbr_partition[i].Part_start)){
				porcentaje1:= Percent(int64(hdrive. Mbr_partition[i].Part_start)-emptySpace,tamañoR);
				s1:= fmt.Sprintf("%.2f", porcentaje1) 
				grafo=grafo+"<TD>LIBRE<br/>"+s1+"%"+"</TD>"
			}
			n:= bytes.IndexByte(hdrive.Mbr_partition[i].Part_name[:], 0)
			grafo=grafo+"<TD>"+string(hdrive.Mbr_partition[i].Part_name[:n])+"<br/>"+string(hdrive. Mbr_partition[i].Part_type)+" "+s+"%"+"</TD>";
			if(i==checklast&&hdrive.Mbr_tamano!=hdrive.Mbr_partition[i].Part_start+hdrive.Mbr_partition[i].Part_size){
				porcentaje:= Percent(int64(hdrive.Mbr_tamano-(hdrive. Mbr_partition[i].Part_start+hdrive. Mbr_partition[i].Part_size)),tamañoR);
				s = fmt.Sprintf("%.2f", porcentaje) 
				grafo=grafo+"<TD>LIBRE<br/> "+s+"%"+"</TD>"
			}
			emptySpace=int64(hdrive. Mbr_partition[i].Part_size)+int64(hdrive. Mbr_partition[i].Part_start);
		}

	}
	grafo=grafo+"</TR></TABLE>> ];\n}";
	dot:=strings.ReplaceAll(path,"\"","");
	f, err := os.Create(strings.ReplaceAll("/"+dot,".png",""))
    if err != nil {
        fmt.Println(err)
        return
    }
    l, err := f.WriteString(grafo)
    if err != nil {
        fmt.Println(err)
        f.Close()
        return
    }
    fmt.Println(l, "bytes written successfully")
    err = f.Close()
    if err != nil {
        fmt.Println(err)
        return
	}
	comando:="C:/Program Files/Graphviz 2.44.1/bin/dot test -Tpng -O"
	cmd:=exec.Command(comando);
	stdout, err := cmd.Output()

    if err != nil {
        fmt.Println(err.Error())
        return
	}
	fmt.Print(string(stdout))
}

func Percent(percent int64, all int64) float64 {
	return ( float64(100) * float64(percent)) /(float64(all))
}