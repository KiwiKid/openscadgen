include <BOSL2/std.scad>;

$fn = 100;

wedgeLength = 46;

wedgeSize = 13;
screwGap1 =35/2;
screwGap2 =7;

module hex(h=4, r=5, chamfer=2){
intersection(){
linear_extrude(h=h)
    hexagon(r=r, anchor=CENTER);
    
    cyl(r=r, h=h, chamfer2=chamfer, anchor=BOT);
    }
    
}

module screwHoles(screwGap=2, screwRadius=1.8, screwChamfer=-2, topFwdExtra=0, screwLength=wedgeSize/2+0.01){

    left(screwGap)
     hex(h=screwLength, r=screwRadius, chamfer=screwChamfer);
   // cyl(r=screwRadius, h=screwLength, chamfer2=screwChamfer, anchor=BOT)
   //  fwd(topFwdExtra)
      //  cyl(r=screwRadius+2, h=10, anchor=BOT);
    
 left(-screwGap)
   // cyl(r=screwRadius, h=screwLength, chamfer2=screwChamfer, anchor=BOT)
   hex(h=screwLength, r=screwRadius, chamfer=screwChamfer);
     //attach(TOP)
    // #fwd(topFwdExtra)
        //cyl(r=screwRadius+2, h=10, anchor=BOT);
        
}

module magnetic_catch_mount(){
difference(){
hull(){
	wedge([wedgeLength,wedgeSize,wedgeSize], anchor=CENTER);
    xcyl(r=3, h=wedgeLength);
    }
       down(10)
    screwHoles(screwGap=screwGap1, screwRadius=1.2, screwLength=10, screwChamfer=1, topFwdExtra=30);
    
    fwd(10)
    xrot(-90)
    screwHoles(screwGap=screwGap2, screwLength=20);
    }
}

magnetic_catch_mount();
//hex();