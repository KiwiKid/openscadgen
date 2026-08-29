

	include <BOSL2/std.scad>;
include <BOSL2/joiners.scad>;
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
    
    
    partType = "plug"; //"plug"; // phoneMount cordHolder wallMount // all-car
    bottomShelf = "includeBottomShelf"; //includeBottomShelf
    phoneHolderWidth = 40;
    phoneHolderHeight = 95;
    phoneSize = [phoneHolderHeight, phoneHolderWidth, 70];
    
    phoneCutoutHeight = 190;
    phoneCutoutSize = [92, 31, phoneCutoutHeight];
    
    horzPhoneCutoutMove = [-phoneHolderHeight/2,0,0];
    
    
    phoneCutout2Size = [75, 32, 80];
    cutoutMove = [0,-2,8];
    
    wallSize = 2;
    phoneRotateZ = -22;
    phoneRotateX = -15;
    phoneRotate = [phoneRotateX,0,phoneRotateZ];
    
    holderCubeHeight = 95;
    holderCubeWidth = 68;
    holderCubeDepth = 31;
    holderCubeSize = [holderCubeWidth, holderCubeHeight, holderCubeDepth];
    holderHoleSize = [holderCubeWidth*0.90, holderCubeHeight*0.85, holderCubeDepth*0.7];
    phoneHolderCubeMove = [-30,phoneHolderWidth/2+35,0];
    centralCordCutoutSize = [35,30,phoneHolderHeight*2];
    
    
    //Slide 
    slideJoinSlide = 100;
    
    slideJoinRotateZ = phoneRotateZ;
   slideJoinRotateX = 90+phoneRotateX;

    
    
    dovetailHeight = 10;
    cutDepthPosition = dovetailHeight+2.1;
    slideBlockDepth = 500;
    slideBlockHeight =150;
    
    slideBlockWidth =200;
    slideOffset = 30;
    slideLargeWidth = 50;
    
        
    dovetailSlope = 30;
    
    cordHoleSize = 5.1;
    cordHoleScale = [1,2,1];
    
    

	module rav4_phone_holder(){
		move([0,-20,0])
        difference(){
        
        
        union(){
        rotate(phoneRotate)
        cuboid(phoneSize, rounding=8);
        
        
            shelfHeight = 35;
            
             if(bottomShelf == "includeBottomShelf"){
                rotate(phoneRotate)
                move(cutoutMove-[0,-5,holderCubeHeight/2+3])
                fwd(10)
                cuboid([95,35,shelfHeight], rounding=3);
            }
        
        difference(){
        move(phoneHolderCubeMove)
        cuboid(holderCubeSize, rounding=10);
        
        if(partType == "plug"){
                    #move(phoneHolderCubeMove)
            cuboid(holderHoleSize, rounding=10);
            }
        }
        

        
        }
        
        
        // charger cord holder hole
        if(partType != "cordHolder"){
            move([-55,25,0])
            scale(cordHoleScale)
            yrot(-5)
            cyl(r=cordHoleSize, h=80);
        }
        
          
        rotate(phoneRotate)
        up(phoneCutoutHeight/2-40)
       move(cutoutMove)
		cuboid(phoneCutoutSize, rounding=3);
        
        
        
        // Side phone cutout
     rotate(phoneRotate)
       rotate([0,90,0])
        move(horzPhoneCutoutMove)
        move(cutoutMove)
		cuboid(phoneCutoutSize, rounding=3);
        
        
        // Cutout to show screen
        rotate(phoneRotate)
        move(cutoutMove)
        fwd(10)
		cuboid(phoneCutout2Size, rounding=3);
        
        
        // Shorten the middle holder struct
        rotate(phoneRotate)
        move(cutoutMove-[40,10,-30])
        fwd(10)
		cuboid([20,30,60], rounding=3);
        
        
        // Central plugged in cord hole
          rotate(phoneRotate)
     //   move(cutoutMove)

        fwd(10)
		cuboid(centralCordCutoutSize, rounding=2);
        
        if (bottomShelf == "includeBottomShelf"){
        // shelf cutout
            rotate(phoneRotate+[-8,0,0])
            move(cutoutMove-[0,-7,holderCubeHeight/2+8])
            fwd(10)
            cuboid([100,30,20], rounding=3);
        
        }
        }
        
        
      /*  move([0,phoneHolderWidth/2,0])
        rotate([0,90,90])
        prismoid(size1=[35,60], size2=[30,55], h=60, rounding=4);*/
       
        
        
	}
    
    
    module dovetailJoiner(type="male"){
             dovetail(type, slide=slideJoinSlide, width=18, height=dovetailHeight, back_width=slideLargeWidth, angle=20);
             }
    
    
    wallMountWidth = 25;
    wallMountSize = [100,60,wallMountWidth];
    screwOffset = 70;
    module screwHole(){
    
      //  yrot(90)
      up(dovetailHeight/2)
        cyl(r=3.5, h=wallMountWidth-dovetailHeight+0.001, chamfer1=-2);
    
    }

    module sideJoin(gender){
        left(slideOffset)
        down(cutDepthPosition)
        if(gender=="male"){
        
            rotate([0,180,0])
            down(dovetailHeight)
            dovetailJoiner("male");
         
        up(slideBlockDepth/2+dovetailHeight)
         cuboid([slideBlockWidth,slideBlockHeight,slideBlockDepth]);
                
        }else if(gender == "female"){
        
        down(slideBlockDepth/2-dovetailHeight+0.1)
          diff("remove")
        cuboid([slideBlockWidth,slideBlockHeight,slideBlockDepth])
          tag("remove") attach(TOP) 
          dovetailJoiner("female");
        }
    
    }


    sliced(renderType=renderType) {

        if(partType == "plug" || partType == "all" || partType == "all-car"){
        intersection(){        
    
       zrot(slideJoinRotateZ)
        xrot(slideJoinRotateX)
            sideJoin("female");
            rav4_phone_holder();
        }
        }
        
        if(partType == "wallMount" || partType == "all"){
            intersection(){
                sideJoin("male");  
                left(30)
                difference(){
                    cuboid(wallMountSize);
                    union(){
                        right(screwOffset)
                        screwHole();
                        left(screwOffset)
                        screwHole();
                        }
                    }
                    
                    }
                
            }
        
        if(partType == "phoneMount" || partType == "all" || partType == "all-car"){
        intersection(){
          
            
       zrot(slideJoinRotateZ)
        xrot(slideJoinRotateX)
            sideJoin("male");
            rav4_phone_holder();
            }
        }
        
        if(partType == "joined"){ 
            rav4_phone_holder();
        }
        
        if(partType == "cordHolder" || partType == "all"  || partType == "all-car"){
        cordHolderHeight = 10;
        cordSize = 3;
        cutSize = 15;
        cordHolderFitReduction = 0.95;
        difference(){
        intersection(){
        
            rav4_phone_holder();
//            scale(cordHoleScale)
       //     yrot(-5)
        /*   xscale(cordHolderFitReduction)
            yscale(cordHolderFitReduction)
            #cyl(r=cordHoleSize, h=cordHolderHeight);
          */  
            
            
            move([-55,5,0])
            scale(cordHoleScale)
            yrot(-5)
            cyl(r=cordHoleSize, h=80);
            }
            
            
            cutoutLength = 25;
            move([-55,5,0])
            scale(cordHoleScale)
            
            down(cutoutLength/4)
            cyl(r=cordSize, h=cutoutLength);
            
            
            move([-55,5,0])
            left(cutSize/2)
            down(cutoutLength/4)
            cuboid([cutSize,5,cutoutLength]);
            }
        }
        
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
                cuboid([sliceSize, sliceSize, sliceThickness], anchor=[-1,-1,-1]);
        } else {
            intersection() {
                children();
                translate(horzSlicePos)
                    cuboid([sliceSize, sliceSize, sliceThickness], anchor=[-1,-1,-1]);
            }
        }
    }

    module vert_slice(raw=false) {
        if (raw) {
            translate(vertSlicePos)
                cuboid([sliceThickness, sliceSize, sliceSize], anchor=[-1,-1,-1]);
        } else {
            intersection() {
                children();
                translate(vertSlicePos)
                    cuboid([sliceThickness, sliceSize, sliceSize], anchor=[-1,-1,-1]);
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

