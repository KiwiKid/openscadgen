

	include <BOSL2/std.scad>;

	$fa = .01;
	$fs = $preview ? 5 : 1;
	$fn = 200;

	/*
	renderType:
	use to print a test slice and confirm sizing before printing:
	 - "horzSlice" - horizontal slices (default)
	 - "vertSlice" - vertical slices
	 - "all" - the whole object
	*/
	renderType = "obj";
    
    phoneHolderWidth = 40;
    phoneHolderHeight = 95;
    phoneSize = [phoneHolderHeight, phoneHolderWidth, 70];
    
    phoneCutoutHeight = 190;
    phoneCutoutSize = [92, 30, phoneCutoutHeight];
    
    horzPhoneCutoutMove = [-phoneHolderHeight/2,0,0];
    
    
    phoneCutout2Size = [80, 32, 80];
    cutoutMove = [0,2,8];
    
    wallSize = 2;
    phoneRotate = [-15,0,-20];
    
    holderCubeSize = [65, 90, 30];
    
    phoneHolderCubeMove = [-30,phoneHolderWidth/2+35,0];

	module rav4_phone_holder(){
		move([0,-20,0])
        difference(){
        
        
        union(){
        rotate(phoneRotate)
        cuboid(phoneSize, rounding=8);
        
        move(phoneHolderCubeMove)
        difference(){
        cuboid(holderCubeSize, rounding=10);
        
        back(38)
        scale([0.85,1.2, 0.8])
        cuboid(holderCubeSize, rounding=10);
        }
        }
        
          
        rotate(phoneRotate)
        up(phoneCutoutHeight/2-40)
       move(cutoutMove)
		cuboid(phoneCutoutSize, rounding=3);
        
        
     rotate(phoneRotate)
       rotate([0,90,0])
        move(horzPhoneCutoutMove)
        move(cutoutMove)
		cuboid(phoneCutoutSize, rounding=3);
        
        
        rotate(phoneRotate)
        move(cutoutMove)
        fwd(10)
		cuboid(phoneCutout2Size, rounding=3);
        
        
          rotate(phoneRotate)
     //   move(cutoutMove)
        fwd(10)
		cuboid([40,25,100], rounding=2);
        
        
        
        }
        
      /*  move([0,phoneHolderWidth/2,0])
        rotate([0,90,90])
        prismoid(size1=[35,60], size2=[30,55], h=60, rounding=4);*/
       
        
        
	}


    sliced(renderType=renderType) {
        rav4_phone_holder();
    }
       








	
     
module sliced(
    renderType = "horzSlice",        // "horzSlice", "vertSlice", or "all"
    sliceSize = 1000,
    sliceThickness = 0.2,
    showRawSlices = false,
    horzSlicePos = [-500, -500, 0],
    vertSlicePos = [0, -500, -500]
) {
   
    module horz_slice(raw=false) {
        if (raw) {
            translate(horzSlicePos)
                cube([sliceSize, sliceSize, sliceThickness], center=false);
        } else {
            intersection() {
                children();
                translate(horzSlicePos)
                    cube([sliceSize, sliceSize, sliceThickness], center=false);
            }
        }
    }

    module vert_slice(raw=false) {
        if (raw) {
            translate(vertSlicePos)
                cube([sliceThickness, sliceSize, sliceSize], center=false);
        } else {
            intersection() {
                children();
                translate(vertSlicePos)
                    cube([sliceThickness, sliceSize, sliceSize], center=false);
            }
        }
    }

    if (renderType == "horzSlice") {
        horz_slice(raw=showRawSlices){
            children();
        }
    } else if (renderType == "vertSlice") {
        vert_slice(raw=showRawSlices){
            children();
        }
    } else if (renderType == "all") {
        // show raw slices for reference
        horz_slice(raw=true);
        vert_slice(raw=true);
        // show full object
        children();
    } else {
        // show full object
        children();
    }
}

